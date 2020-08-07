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
