/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/Azure/k8s-infra/hack/generator/pkg/jsonast"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"

	"k8s.io/klog/v2"
)

type CodeGenerator struct {
	configuration *Configuration
}

func NewCodeGenerator(configurationFile string) (*CodeGenerator, error) {
	config, err := loadConfiguration(configurationFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration file '%v' (%w)", configurationFile, err)
	}

	err = config.Validate()
	if err != nil {
		return nil, fmt.Errorf("configuration loaded from '%v' is invalid (%w)", configurationFile, err)
	}

	result := &CodeGenerator{configuration: config}

	return result, nil
}

func (generator *CodeGenerator) Generate(ctx context.Context, outputFolder string) error {

	klog.V(0).Infof("Loading JSON schema %v", generator.configuration.SchemaURL)
	schema, err := loadSchema(generator.configuration.SchemaURL)
	if err != nil {
		return fmt.Errorf("error loading schema from '%v' (%w)", generator.configuration.SchemaURL, err)
	}

	klog.V(0).Infof("Cleaning output folder '%v'", outputFolder)
	err = cleanFolder(outputFolder)
	if err != nil {
		return fmt.Errorf("error cleaning output folder '%v' (%w)", generator.configuration.SchemaURL, err)
	}

	scanner := jsonast.NewSchemaScanner(astmodel.NewIdentifierFactory())

	klog.V(0).Infof("Walking JSON schema")
	defs, err := scanner.GenerateDefinitions(ctx, schema.Root())
	if err != nil {
		return fmt.Errorf("failed to walk JSON schema (%w)", err)
	}

	packages, err := generator.CreatePackagesForDefinitions(defs)
	if err != nil {
		return fmt.Errorf("failed to assign generated definitions to packages (%w)", err)
	}

	doReport(packages)

	fileCount := 0
	definitionCount := 0

	// emit each package
	klog.V(0).Infof("Writing output files into %v", outputFolder)
	for _, pkg := range packages {

		// create directory if not already there
		outputDir := filepath.Join(outputFolder, pkg.GroupName, pkg.PackageName)
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			klog.V(5).Infof("Creating directory '%s'\n", outputDir)
			err = os.MkdirAll(outputDir, 0700)
			if err != nil {
				klog.Fatalf("Unable to create directory '%s'", outputDir)
			}
		}

		count, err := pkg.EmitDefinitions(outputDir)
		if err != nil {
			return fmt.Errorf("error writing definitions into '%v' (%w)", outputDir, err)
		}

		fileCount += count
		definitionCount += pkg.DefinitionCount()
	}

	klog.V(0).Infof("Completed writing %v files containing %v definitions", fileCount, definitionCount)

	return nil
}

func doReport(defs []*astmodel.PackageDefinition) {

	// groupName → ordered (by version) list
	newDefs := make(map[string][]*astmodel.PackageDefinition)

	for _, def := range defs {
		if group, ok := newDefs[def.GroupName]; ok {
			newDefs[def.GroupName] = append(group, def)
		} else {
			newDefs[def.GroupName] = []*astmodel.PackageDefinition{def}
		}
	}

	for _, group := range newDefs {
		sort.Slice(group, func(i, j int) bool {
			return group[i].PackageName < group[j].PackageName
		})

		reportGroup(group)
	}
}

func reportGroup(versions []*astmodel.PackageDefinition) {

	fmt.Printf("Group: %s\n", versions[0].GroupName)

	types := make(map[string][]astmodel.TypeDefiner)
	for _, version := range versions {
		for _, T := range version.Definitions() {
			typeName := T.Name().Name()
			if Tversions, ok := types[typeName]; ok {
				types[typeName] = append(Tversions, T)
			} else {
				types[typeName] = []astmodel.TypeDefiner{T}
			}
		}
	}

	for _, typeVersions := range types {
		printedType := false
		printTypeIfNeeded := func() {
			if !printedType {
				printedType = true
				fmt.Printf(" - comparing type %s\n", typeVersions[0].Name().Name())
			}
		}

		for i := range typeVersions {
			if i+1 >= len(typeVersions) {
				break
			}

			oldType := typeVersions[i]
			newType := typeVersions[i+1]

			compareTypes(oldType, newType, printTypeIfNeeded)
		}
	}
}

func findType(oldType astmodel.TypeDefiner, newTypes []astmodel.TypeDefiner) astmodel.TypeDefiner {
	for _, newType := range newTypes {
		if oldType.Name().Name() == newType.Name().Name() {
			return newType
		}
	}

	return nil
}

func compareTypes(oldType astmodel.TypeDefiner, newType astmodel.TypeDefiner, printTypeIfNeeded func()) {

	printedVersion := false
	printVersionIfNeeded := func() {
		printTypeIfNeeded()
		if !printedVersion {
			printedVersion = true
			fmt.Printf("   - version %s to %s\n", oldType.Name().PackageReference.PackageName(), newType.Name().PackageReference.PackageName())
		}
	}

	// only compare structs for now
	if oldStruct, ok := oldType.(*astmodel.StructDefinition); ok {
		if newStruct, ok := newType.(*astmodel.StructDefinition); ok {

			for _, oldField := range oldStruct.StructType.Fields() {

				newField := findField(oldField, newStruct.StructType.Fields())
				if newField == nil {
					printVersionIfNeeded()
					fmt.Printf("      ! field removed: %s\n", oldField.FieldName())
				} else {
					if !astmodel.TypesEqualIgnoringVersions(oldField.FieldType(), newField.FieldType()) {
						if optional, ok := newField.FieldType().(*astmodel.OptionalType); ok &&
							astmodel.TypesEqualIgnoringVersions(oldField.FieldType(), optional.ElementType()) {
							printVersionIfNeeded()
							fmt.Printf("      + field became optional: %s\n", newField.FieldName())
						} else {
							printVersionIfNeeded()

							fset := token.NewFileSet()
							fset.AddFile("text", 1, 102400)

							var buffer bytes.Buffer
							format.Node(&buffer, fset, oldField.FieldType().AsType())

							var buffer2 bytes.Buffer
							format.Node(&buffer2, fset, newField.FieldType().AsType())

							fmt.Printf("      ! field type changed: %s (was %s, now %s)\n", oldField.FieldName(), buffer.String(), buffer2.String())
						}
					}
				}
			}

			for _, newField := range newStruct.StructType.Fields() {
				oldField := findField(newField, oldStruct.StructType.Fields())
				if oldField == nil {
					printVersionIfNeeded()
					fmt.Printf("      + field added: %s\n", newField.FieldName())
				}
			}
		}
	}
}

func findField(oldField *astmodel.FieldDefinition, newFields []*astmodel.FieldDefinition) *astmodel.FieldDefinition {
	for _, newField := range newFields {
		if oldField.FieldName() == newField.FieldName() {
			return newField
		}
	}

	return nil
}

func (generator *CodeGenerator) CreatePackagesForDefinitions(definitions []astmodel.TypeDefiner) ([]*astmodel.PackageDefinition, error) {
	packages := make(map[astmodel.PackageReference]*astmodel.PackageDefinition)
	for _, def := range definitions {

		shouldExport, reason := generator.configuration.ShouldExport(def)
		defName := def.Name()
		groupName, pkgName, err := defName.PackageReference.GroupAndPackage()
		if err != nil {
			return nil, err
		}

		switch shouldExport {
		case Skip:
			klog.V(2).Infof("Skipping %s/%s because %s", groupName, pkgName, reason)

		case Export:
			if reason == "" {
				klog.V(3).Infof("Exporting %s/%s", groupName, pkgName)
			} else {
				klog.V(2).Infof("Exporting %s/%s because %s", groupName, pkgName, reason)
			}

			pkgRef := defName.PackageReference
			if pkg, ok := packages[pkgRef]; ok {
				pkg.AddDefinition(def)
			} else {
				pkg = astmodel.NewPackageDefinition(groupName, pkgName)
				pkg.AddDefinition(def)
				packages[pkgRef] = pkg
			}
		}
	}

	var pkgs []*astmodel.PackageDefinition
	for _, pkg := range packages {
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

func loadConfiguration(configurationFile string) (*Configuration, error) {
	data, err := ioutil.ReadFile(configurationFile)
	if err != nil {
		return nil, err
	}

	result := NewConfiguration()

	err = yaml.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func loadSchema(source string) (*gojsonschema.Schema, error) {
	sl := gojsonschema.NewSchemaLoader()
	schema, err := sl.Compile(gojsonschema.NewReferenceLoader(source))
	if err != nil {
		return nil, fmt.Errorf("error loading schema from '%v' (%w)", source, err)
	}

	return schema, nil
}

//TODO: Only clean generated files
func cleanFolder(outputFolder string) error {
	err := os.RemoveAll(outputFolder)
	if err != nil {
		return fmt.Errorf("error removing output folder '%v' (%w)", outputFolder, err)
	}

	err = os.Mkdir(outputFolder, 0700)
	if err != nil {
		return fmt.Errorf("error creating output folder '%v' (%w)", outputFolder, err)
	}

	return nil
}
