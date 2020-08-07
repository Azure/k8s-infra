/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/Azure/k8s-infra/hack/generator/pkg/config"
	"github.com/Azure/k8s-infra/hack/generator/pkg/jsonast"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// augmentResourcesWithStatus creates a PipelineStage to add Swagger information into the 'status' part of CRD
func augmentResourcesWithStatus(idFactory astmodel.IdentifierFactory, config *config.Configuration) PipelineStage {
	return PipelineStage{
		"Add information from Swagger specs for 'status' fields",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			if config.Status.SchemaRoot == "" {
				klog.Warningf("no status schema root specified, will not generate status types")
				return types, nil
			}

			klog.V(2).Info("loading Swagger data")

			swaggerTypes, err := loadSwaggerData(ctx, idFactory, config)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to load Swagger data")
			}

			klog.Infof("loaded Swagger data (%v resources, %v other types)", len(swaggerTypes.resources), len(swaggerTypes.types))

			newTypes := make(astmodel.Types)

			resourceLookup, otherTypes := makeStatusLookup(swaggerTypes)
			for _, otherType := range otherTypes {
				newTypes.Add(otherType)
			}

			found := 0
			notFound := 0
			for typeName, typeDef := range types {
				if resource, ok := typeDef.Type().(*astmodel.ResourceType); ok {
					if statusDef, ok := resourceLookup[typeName]; ok {
						//klog.Infof("Found swagger information for %v", typeName)
						newTypes.Add(astmodel.MakeTypeDefinition(typeName, resource.WithStatus(statusDef)))
						found++
					} else {
						//klog.Warningf("No swagger information found for %v", typeName)
						newTypes.Add(typeDef)
						notFound++
					}
				} else {
					newTypes.Add(typeDef)
				}
			}

			klog.Infof("Found status information for %v resources", found)
			klog.Infof("Missing status information for %v resources", notFound)
			klog.Infof("Input %v types, output %v types", len(types), len(newTypes))

			return newTypes, nil
		},
	}
}

type resourceLookup map[astmodel.TypeName]astmodel.Type

// makeStatusLookup makes a map of old name (non-status name) to status type
func makeStatusLookup(swaggerTypes swaggerTypes) (resourceLookup, []astmodel.TypeDefinition) {
	statusVisitor := makeStatusVisitor()

	var otherTypes []astmodel.TypeDefinition
	for typeName, typeDef := range swaggerTypes.types {
		newName := appendStatusToName(typeName)
		otherTypes = append(otherTypes, astmodel.MakeTypeDefinition(newName, statusVisitor.Visit(typeDef.Type(), nil)))
	}

	resources := make(resourceLookup)
	for resourceName, resourceDef := range swaggerTypes.resources {
		resources[resourceName] = statusVisitor.Visit(resourceDef.Type(), nil)
	}

	return resources, otherTypes
}

func appendStatusToName(typeName astmodel.TypeName) astmodel.TypeName {
	return astmodel.MakeTypeName(typeName.PackageReference, typeName.Name()+"_Status")
}

func makeStatusVisitor() astmodel.TypeVisitor {
	visitor := astmodel.MakeTypeVisitor()

	visitor.VisitTypeName = func(this *astmodel.TypeVisitor, it astmodel.TypeName, ctx interface{}) astmodel.Type {
		return appendStatusToName(it)
	}

	return visitor
}

var swaggerVersionRegex = regexp.MustCompile("\\d{4}-\\d{2}-\\d{2}(-preview)?")

type swaggerTypes struct {
	resources astmodel.Types
	types     astmodel.Types
}

var skipDirectories = []string{
	"/examples/",
	"/quickstart-templates/",
	"/control-plane/",
	"/data-plane/",
}

func loadSwaggerData(ctx context.Context, idFactory astmodel.IdentifierFactory, config *config.Configuration) (swaggerTypes, error) {

	result := swaggerTypes{
		resources: make(astmodel.Types),
		types:     make(astmodel.Types),
	}

	cache := jsonast.MakeOpenAPISchemaCache()

	schemas, err := loadAllSchemas(ctx, config.Status.SchemaRoot, cache)
	if err != nil {
		return swaggerTypes{}, err
	}

	for schemaPath, schema := range schemas {
		// these have already been tested in the loadAllSchemas function
		outputGroup := swaggerGroupRegex.FindString(schemaPath)
		outputVersion := swaggerVersionRegex.FindString(schemaPath)

		// see if there is a config override for the namespace (group) name
		for _, configSchema := range config.Status.Schemas {
			configSchemaPath := path.Join(config.Status.SchemaRoot, configSchema.BasePath)
			if strings.HasPrefix(schemaPath, configSchemaPath) {
				// found a corresponding schema
				if configSchema.Suffix != "" {
					outputGroup += "." + configSchema.Suffix
				}

				break
			}
		}

		extractor := typeExtractor{
			outputVersion: outputVersion,
			outputGroup:   outputGroup,
			idFactory:     idFactory,
			cache:         cache,
			config:        config,
		}

		err := extractor.extractTypes(ctx, schemaPath, schema, result.resources, result.types)
		if err != nil {
			return swaggerTypes{}, err
		}
	}

	return result, nil
}

func loadAllSchemas(
	ctx context.Context,
	rootPath string,
	cache *jsonast.OpenAPISchemaCache) (map[string]spec.Swagger, error) {

	var wg sync.WaitGroup

	var mutex sync.Mutex
	schemas := make(map[string]spec.Swagger)
	var sharedErr error

	err := filepath.Walk(rootPath, func(filePath string, fileInfo os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		for _, skipDir := range skipDirectories {
			if strings.Contains(filePath, skipDir) {
				return filepath.SkipDir // magic error
			}
		}

		if !fileInfo.IsDir() &&
			filepath.Ext(filePath) == ".json" &&
			swaggerGroupRegex.MatchString(filePath) &&
			swaggerVersionRegex.MatchString(filePath) {

			wg.Add(1)
			go func() {
				swagger, err := cache.PreloadCache(filePath)
				mutex.Lock()
				if err != nil {
					// first error wins
					if sharedErr == nil {
						sharedErr = err
					}
				} else {
					schemas[filePath] = swagger
				}
				mutex.Unlock()

				wg.Done()
			}()
		}

		return nil
	})

	wg.Wait()

	if err != nil {
		return nil, err
	}

	if sharedErr != nil {
		return nil, sharedErr
	}

	return schemas, nil
}

type typeExtractor struct {
	idFactory astmodel.IdentifierFactory
	config    *config.Configuration
	cache     *jsonast.OpenAPISchemaCache
	// group for output types (e.g. Microsoft.Network.Frontdoor)
	outputGroup   string
	outputVersion string
}

func (extractor *typeExtractor) extractTypes(
	ctx context.Context,
	filePath string,
	swagger spec.Swagger,
	resources astmodel.Types,
	types astmodel.Types) error {

	packageName := extractor.idFactory.CreatePackageNameFromVersion(extractor.outputVersion)

	scanner := jsonast.NewSchemaScanner(extractor.idFactory, extractor.config)

	for operationPath, op := range swagger.Paths.Paths {
		put := op.Put
		if put == nil {
			continue
		}

		resourceName, err := extractor.resourceNameFromOperationPath(packageName, operationPath)
		if err != nil {
			klog.Errorf("unable to produce resource name for path %q (in %s): %v", operationPath, filePath, err)
			continue
		}

		resourceType, err := extractor.resourceTypeFromOperation(ctx, scanner, swagger, filePath, put)
		if err != nil {
			if err == context.Canceled {
				return err
			}

			klog.Errorf("unable to produce type for %v (in %s): %v", resourceName, filePath, err)
			continue // TODO: make fatal
		}

		if resourceType != nil {
			if existingResource, ok := resources[resourceName]; ok {
				if !astmodel.TypeEquals(existingResource.Type(), resourceType) {
					klog.Errorf("RESOURCE already defined differently 😱: %v (%s)", resourceName, filePath)
				}
			} else {
				resources.Add(astmodel.MakeTypeDefinition(resourceName, resourceType))
			}
		}
	}

	for _, def := range scanner.Definitions() {
		// get generated-aside definitions too
		if existingDef, ok := types[def.Name()]; ok {
			if !astmodel.TypeEquals(existingDef.Type(), def.Type()) {
				klog.Errorf("type already defined differently: %v (%s)", def.Name(), filePath)
			}
		} else {
			types.Add(def)
		}
	}

	return nil
}

func (extractor *typeExtractor) resourceNameFromOperationPath(packageName string, operationPath string) (astmodel.TypeName, error) {
	// infer name from path
	_, name, err := inferNameFromURLPath(operationPath)
	if err != nil {
		return astmodel.TypeName{}, errors.Wrapf(err, "unable to infer name from path")
	}

	packageRef := astmodel.MakeLocalPackageReference(extractor.idFactory.CreateGroupName(extractor.outputGroup), packageName)
	return astmodel.MakeTypeName(packageRef, name), nil
}

func (extractor *typeExtractor) resourceTypeFromOperation(
	ctx context.Context,
	scanner *jsonast.SchemaScanner,
	schemaRoot spec.Swagger,
	filePath string,
	operation *spec.Operation) (astmodel.Type, error) {

	for _, param := range operation.Parameters {
		if param.In == "body" && param.Required { // assume this is the Resource
			schema := jsonast.MakeOpenAPISchema(
				*param.Schema,
				schemaRoot,
				filePath,
				extractor.outputGroup,
				extractor.outputVersion,
				extractor.cache)

			return scanner.RunHandlerForSchema(ctx, schema)
		}
	}

	return nil, nil
}

func inferNameFromURLPath(operationPath string) (string, string, error) {

	group := ""
	name := ""

	urlParts := strings.Split(operationPath, "/")
	reading := false
	skippedLast := false
	for _, urlPart := range urlParts {
		if reading {
			if len(urlPart) > 0 && urlPart[0] != '{' {
				name += strings.ToUpper(urlPart[0:1]) + urlPart[1:]
				skippedLast = false
			} else {
				if skippedLast {
					// this means two {parameters} in a row
					return "", "", errors.Errorf("multiple parameters in path")
				}

				skippedLast = true
			}
		} else if swaggerGroupRegex.MatchString(urlPart) {
			group = urlPart
			reading = true
		}
	}

	if reading == false {
		return "", "", errors.Errorf("no group name (‘Microsoft…’ = %q) found", group)
	}

	if name == "" {
		return "", "", errors.Errorf("couldn’t infer name")
	}

	return group, name, nil
}

// based on: https://github.com/Azure/autorest/blob/85de19623bdce3ccc5000bae5afbf22a49bc4665/core/lib/pipeline/metadata-generation.ts#L25
var swaggerGroupRegex = regexp.MustCompile("[Mm]icrosoft\\.[^/\\\\]+")
