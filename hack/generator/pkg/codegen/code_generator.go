/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonreference"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/Azure/k8s-infra/hack/generator/pkg/config"
	"github.com/Azure/k8s-infra/hack/generator/pkg/jsonast"
)

// CodeGenerator is a generator of code
type CodeGenerator struct {
	configuration *config.Configuration
}

// NewCodeGenerator produces a new Generator with the given configuration
func NewCodeGenerator(configurationFile string) (*CodeGenerator, error) {
	config, err := loadConfiguration(configurationFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load configuration file %q", configurationFile)
	}

	err = config.Initialize()
	if err != nil {
		return nil, errors.Wrapf(err, "configuration loaded from %q is invalid", configurationFile)
	}

	result := &CodeGenerator{configuration: config}

	return result, nil
}

// Generate produces the Go code corresponding to the configured JSON schema in the given output folder
func (generator *CodeGenerator) Generate(ctx context.Context) error {
	klog.V(1).Infof("Generator version: %v", combinedVersion())
	klog.V(0).Infof("Loading JSON schema %v", generator.configuration.SchemaURL)
	schema, err := loadSchema(ctx, generator.configuration.SchemaURL)
	if err != nil {
		return errors.Wrapf(err, "error loading schema from %q", generator.configuration.SchemaURL)
	}

	scanner := jsonast.NewSchemaScanner(astmodel.NewIdentifierFactory(), generator.configuration)

	klog.V(0).Infof("Walking JSON schema")

	defs, err := scanner.GenerateDefinitions(ctx, schema.Root())
	if err != nil {
		return errors.Wrapf(err, "failed to walk JSON schema")
	}

	pipeline := []PipelineStage{
		applyExportFilters(generator.configuration),
		deleteGeneratedCode(ctx, generator.configuration.OutputPath),
	}

	for i, stage := range pipeline {
		klog.V(0).Infof("Pipeline stage %d/%d: %s", i+1, len(pipeline), stage.Name)
		defs, err = stage.Action(defs)
		if err != nil {
			return errors.Wrapf(err, "Failed during pipeline stage %s", stage.Name)
		}
	}

	roots := astmodel.CollectResourceDefinitions(defs)
	defs, err = astmodel.StripUnusedDefinitions(roots, defs)
	if err != nil {
		return errors.Wrapf(err, "failed to strip unused definitions")
	}

	packages, err := generator.CreatePackagesForDefinitions(defs)
	if err != nil {
		return errors.Wrapf(err, "failed to assign generated definitions to packages")
	}

	packages, err = generator.MarkLatestResourceVersionsForStorage(packages)
	if err != nil {
		return errors.Wrapf(err, "unable to mark latest resource versions for as storage versions")
	}

	fileCount := 0
	definitionCount := 0

	// emit each package
	klog.V(0).Infof("Writing output files into %v", generator.configuration.OutputPath)
	for _, pkg := range packages {
		if ctx.Err() != nil { // check for cancellation
			return ctx.Err()
		}

		// create directory if not already there
		outputDir := filepath.Join(generator.configuration.OutputPath, pkg.GroupName, pkg.PackageName)
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			klog.V(5).Infof("Creating directory %q\n", outputDir)
			err = os.MkdirAll(outputDir, 0700)
			if err != nil {
				klog.Fatalf("Unable to create directory %q", outputDir)
			}
		}

		count, err := pkg.EmitDefinitions(outputDir)
		if err != nil {
			return errors.Wrapf(err, "error writing definitions into %q", outputDir)
		}

		fileCount += count
		definitionCount += pkg.DefinitionCount()
	}

	klog.V(0).Infof("Completed writing %v files containing %v definitions", fileCount, definitionCount)

	return nil
}

// MarkLatestResourceVersionsForStorage marks the latest version of each resource as the storage version
func (generator *CodeGenerator) MarkLatestResourceVersionsForStorage(
	pkgs []*astmodel.PackageDefinition) ([]*astmodel.PackageDefinition, error) {

	var result []*astmodel.PackageDefinition

	resourceLookup, err := groupResourcesByVersion(pkgs)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {

		resultPkg := astmodel.NewPackageDefinition(pkg.GroupName, pkg.PackageName, pkg.GeneratorVersion)
		for _, def := range pkg.Definitions() {
			// see if it is a resource
			if resourceDef, ok := def.(*astmodel.ResourceDefinition); ok {

				unversionedName, err := getUnversionedName(resourceDef.Name())
				if err != nil {
					// should never happen as all resources have versioned names
					return nil, err
				}

				allVersionsOfResource := resourceLookup[unversionedName]
				latestVersionOfResource := allVersionsOfResource[len(allVersionsOfResource)-1]

				thisPackagePath := resourceDef.Name().PackageReference.PackagePath()
				latestPackagePath := latestVersionOfResource.Name().PackageReference.PackagePath()

				// mark as storage version if it's the latest version
				isLatestVersion := thisPackagePath == latestPackagePath
				if isLatestVersion {
					resourceDef = resourceDef.MarkAsStorageVersion()
				}

				resultPkg.AddDefinition(resourceDef)
			} else {
				// otherwise simply add it
				resultPkg.AddDefinition(def)
			}
		}

		result = append(result, resultPkg)
	}

	return result, nil
}

func getUnversionedName(name *astmodel.TypeName) (unversionedName, error) {
	group, _, err := name.PackageReference.GroupAndPackage()
	if err != nil {
		return unversionedName{}, err
	}

	return unversionedName{group, name.Name()}, nil
}

type unversionedName struct {
	group string
	name  string
}

func groupResourcesByVersion(
	pkgs []*astmodel.PackageDefinition) (map[unversionedName][]*astmodel.ResourceDefinition, error) {

	result := make(map[unversionedName][]*astmodel.ResourceDefinition)

	for _, pkg := range pkgs {
		for _, def := range pkg.Definitions() {
			if resourceDef, ok := def.(*astmodel.ResourceDefinition); ok {
				name, err := getUnversionedName(resourceDef.Name())
				if err != nil {
					// this should never happen as resources will all have versioned names
					return nil, errors.Wrapf(err, "Unable to extract unversioned name in groupResources")
				}

				result[name] = append(result[name], resourceDef)
			}
		}
	}

	// order each set of resources by package name (== by version as these are sortable dates)
	for _, slice := range result {
		sort.Slice(slice, func(i, j int) bool {
			return slice[i].Name().PackageReference.PackageName() < slice[j].Name().PackageReference.PackageName()
		})
	}

	return result, nil
}


// CreatePackagesForDefinitions groups type definitions into packages
func (generator *CodeGenerator) CreatePackagesForDefinitions(
	definitions []astmodel.TypeDefiner) ([]*astmodel.PackageDefinition, error) {

	genVersion := combinedVersion()
	packages := make(map[astmodel.PackageReference]*astmodel.PackageDefinition)
	for _, def := range definitions {
		defName := def.Name()
		groupName, pkgName, err := defName.PackageReference.GroupAndPackage()
		if err != nil {
			return nil, err
		}

		pkgRef := defName.PackageReference
		if pkg, ok := packages[pkgRef]; ok {
			pkg.AddDefinition(def)
		} else {
			pkg = astmodel.NewPackageDefinition(groupName, pkgName, genVersion)
			pkg.AddDefinition(def)
			packages[pkgRef] = pkg
		}
	}

	var pkgs []*astmodel.PackageDefinition
	for _, pkg := range packages {
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

func loadConfiguration(configurationFile string) (*config.Configuration, error) {
	data, err := ioutil.ReadFile(configurationFile)
	if err != nil {
		return nil, err
	}

	result := config.NewConfiguration()

	err = yaml.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type cancellableFileSystem struct {
	ctx context.Context
}

var _ http.FileSystem = &cancellableFileSystem{} // interface assertion

func (fs *cancellableFileSystem) Open(source string) (http.File, error) {
	if fs.ctx.Err() != nil { // check for cancellation
		return nil, fs.ctx.Err()
	}

	return os.Open(source)
}

type cancellableJSONLoaderFactory struct {
	ctx   context.Context
	inner gojsonschema.JSONLoaderFactory
}

var _ gojsonschema.JSONLoaderFactory = &cancellableJSONLoaderFactory{}

func (factory *cancellableJSONLoaderFactory) New(source string) gojsonschema.JSONLoader {
	return &cancellableJSONLoader{factory.ctx, factory.inner.New(source)}
}

type cancellableJSONLoader struct {
	ctx   context.Context
	inner gojsonschema.JSONLoader
}

var _ gojsonschema.JSONLoader = &cancellableJSONLoader{}

func (loader *cancellableJSONLoader) LoadJSON() (interface{}, error) {
	if loader.ctx.Err() != nil { // check for cancellation
		return nil, loader.ctx.Err()
	}

	return loader.inner.LoadJSON()
}

func (loader *cancellableJSONLoader) JsonSource() interface{} {
	return loader.inner.JsonSource()
}

func (loader *cancellableJSONLoader) JsonReference() (gojsonreference.JsonReference, error) {
	if loader.ctx.Err() != nil { // check for cancellation
		return gojsonreference.JsonReference{}, loader.ctx.Err()
	}

	return loader.inner.JsonReference()
}

func (loader *cancellableJSONLoader) LoaderFactory() gojsonschema.JSONLoaderFactory {
	return &cancellableJSONLoaderFactory{loader.ctx, loader.inner.LoaderFactory()}
}

func loadSchema(ctx context.Context, source string) (*gojsonschema.Schema, error) {
	sl := gojsonschema.NewSchemaLoader()
	loader := &cancellableJSONLoader{
		ctx,
		gojsonschema.NewReferenceLoaderFileSystem(source, &cancellableFileSystem{ctx}),
	}

	schema, err := sl.Compile(loader)
	if err != nil {
		return nil, errors.Wrapf(err, "error loading schema from %q", source)
	}

	return schema, nil
}
