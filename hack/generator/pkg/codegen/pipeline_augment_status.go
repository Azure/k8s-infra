/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"io/ioutil"
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

/* augmentResourcesWithStatus creates a PipelineStage to add status information into the generated resources.

This information is derived from the Azure Swagger specifications. We parse the Swagger specs and look for
any actions that appear to be ARM resources (have PUT methods with types we can use and appropriate names in the
action path). Then for each resource, we use the existing JSON AST parser to extract the status type
(the type-definition part of swagger is the same as JSON Schema).

Next, we walk over all the resources we are currently generating CRDs for and attempt to locate
a match for the resource in the status information we have parsed. If we locate a match, it is
added to the Status field of the Resource type, after we have renamed all the status types to
avoid any conflicts with existing Spec types that have already been defined.

*/
func augmentResourcesWithStatus(idFactory astmodel.IdentifierFactory, config *config.Configuration) PipelineStage {
	return PipelineStage{
		"augmentStatus",
		"Add information from Swagger specs for 'status' fields",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {

			if config.Status.SchemaRoot == "" {
				klog.Warningf("No status schema root specified, will not generate status types")
				return types, nil
			}

			klog.V(1).Infof("Loading Swagger data from %q", config.Status.SchemaRoot)

			swaggerTypes, err := loadSwaggerData(ctx, idFactory, config)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to load Swagger data")
			}

			klog.V(1).Infof("Loaded Swagger data (%v resources, %v other types)", len(swaggerTypes.resources), len(swaggerTypes.otherTypes))

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
						klog.V(4).Infof("Swagger information found for %v", typeName)
						newTypes.Add(astmodel.MakeTypeDefinition(typeName, resource.WithStatus(statusDef)))
						found++
					} else {
						// TODO: eventually this will be a warning
						klog.V(2).Infof("No swagger information found for %v", typeName)
						newTypes.Add(typeDef)
						notFound++
					}
				} else {
					newTypes.Add(typeDef)
				}
			}

			klog.V(1).Infof("Found status information for %v resources", found)
			klog.V(1).Infof("Missing status information for %v resources", notFound)
			klog.V(1).Infof("Input %v types, output %v types", len(types), len(newTypes))

			return newTypes, nil
		},
	}
}

type resourceLookup map[astmodel.TypeName]astmodel.Type

// makeStatusLookup makes a map of old name (non-status name) to status type
func makeStatusLookup(swaggerTypes swaggerTypes) (resourceLookup, []astmodel.TypeDefinition) {
	statusVisitor := makeStatusVisitor()

	var otherTypes []astmodel.TypeDefinition
	for typeName, typeDef := range swaggerTypes.otherTypes {
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
	resources  astmodel.Types
	otherTypes astmodel.Types
}

var skipDirectories = []string{
	"/examples/",
	"/quickstart-templates/",
	"/control-plane/",
	"/data-plane/",
}

func loadSwaggerData(ctx context.Context, idFactory astmodel.IdentifierFactory, config *config.Configuration) (swaggerTypes, error) {

	result := swaggerTypes{
		resources:  make(astmodel.Types),
		otherTypes: make(astmodel.Types),
	}

	schemas, err := loadAllSchemas(ctx, config.Status.SchemaRoot)
	if err != nil {
		return swaggerTypes{}, err
	}

	cache := jsonast.MakeOpenAPISchemaCache(schemas)

	for schemaPath, schema := range schemas {
		// these have already been tested in the loadAllSchemas function
		outputGroup := swaggerGroupRegex.FindString(schemaPath)
		outputVersion := swaggerVersionRegex.FindString(schemaPath)

		// see if there is a config override for this file
		for _, schemaOverride := range config.Status.Overrides {
			configSchemaPath := path.Join(config.Status.SchemaRoot, schemaOverride.BasePath)
			if strings.HasPrefix(schemaPath, configSchemaPath) {
				// found an override: apply it
				if schemaOverride.Suffix != "" {
					outputGroup += "." + schemaOverride.Suffix
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

		err := extractor.extractTypes(ctx, schemaPath, schema, result.resources, result.otherTypes)
		if err != nil {
			return swaggerTypes{}, err
		}
	}

	return result, nil
}

func loadAllSchemas(
	ctx context.Context,
	rootPath string) (map[string]spec.Swagger, error) {

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
				fileContent, err := ioutil.ReadFile(filePath)

				var swagger spec.Swagger
				if err != nil {
					err = errors.Wrap(err, "unable to read swagger file")
				} else {
					err = swagger.UnmarshalJSON(fileContent)
					if err != nil {
						err = errors.Wrap(err, "unable to parse swagger file")
					}
				}

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
