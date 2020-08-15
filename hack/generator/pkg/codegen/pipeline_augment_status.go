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
	"golang.org/x/sync/errgroup"
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

var swaggerVersionRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}(-preview)?`)

type swaggerTypes struct {
	resources  astmodel.Types
	otherTypes astmodel.Types
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
		// these have already been tested in the loadAllSchemas function so are guaranteed to match
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
			return swaggerTypes{}, errors.Wrapf(err, "error processing %q", schemaPath)
		}
	}

	return result, nil
}

// TODO: is there, perhaps, a way to detect these without hardcoding these paths?
var skipDirectories = []string{
	"/examples/",
	"/quickstart-templates/",
	"/control-plane/",
	"/data-plane/",
}

func shouldSkipDir(filePath string) bool {
	for _, skipDir := range skipDirectories {
		if strings.Contains(filePath, skipDir) {
			return true
		}
	}

	return false
}

// loadAllSchemas walks all .json files in the given rootPath in directories
// of the form "Microsoft.GroupName/…/2000-01-01/…" (excluding those matching
// shouldSkipDir), and returns those files in a map of path→swagger spec.
func loadAllSchemas(
	ctx context.Context,
	rootPath string) (map[string]spec.Swagger, error) {

	var eg errgroup.Group

	var mutex sync.Mutex
	schemas := make(map[string]spec.Swagger)

	err := filepath.Walk(rootPath, func(filePath string, fileInfo os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if shouldSkipDir(filePath) {
			return filepath.SkipDir // this is a magic error
		}

		if !fileInfo.IsDir() &&
			filepath.Ext(filePath) == ".json" &&
			swaggerGroupRegex.MatchString(filePath) &&
			swaggerVersionRegex.MatchString(filePath) {

			// all files are loaded in parallel to speed this up
			eg.Go(func() error {
				var swagger spec.Swagger

				fileContent, err := ioutil.ReadFile(filePath)
				if err != nil {
					return errors.Wrap(err, "unable to read swagger file")
				}

				err = swagger.UnmarshalJSON(fileContent)
				if err != nil {
					return errors.Wrap(err, "unable to parse swagger file")
				}

				mutex.Lock()
				schemas[filePath] = swagger
				mutex.Unlock()

				return nil
			})
		}

		return nil
	})

	egErr := eg.Wait() // for files to finish loading

	if err != nil {
		return nil, err
	}

	if egErr != nil {
		return nil, egErr
	}

	return schemas, nil
}
