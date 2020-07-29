/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"os"
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

			klog.V(2).Info("loading Swagger data")

			statusTypes, err := loadSwaggerData(ctx, idFactory, config)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to load Swagger data")
			}

			klog.Infof("loaded Swagger data (%v types)", len(statusTypes))

			newTypes := make(astmodel.Types)

			statusVisitor := makeStatusVisitor(statusTypes)

			for typeName, typeDef := range types {
				// insert all status objects regardless of if they are used;
				// they will be trimmed by a later phase if they are not
				if statusDef, ok := statusTypes[typeName]; ok {
					statusType := statusDef.Type()
					switch statusType.(type) {
					case *astmodel.ResourceType:
						panic("should never happen")
					case *astmodel.ObjectType:
						statusName := appendStatusToName(typeName)
						statusType = statusVisitor.Visit(statusType, nil)
						newTypes.Add(astmodel.MakeTypeDefinition(statusName, statusType))
					}
				}

				if resource, ok := typeDef.Type().(*astmodel.ResourceType); ok {
					if statusDef, ok := statusTypes[typeName]; ok {
						statusType := statusDef.Type()
						statusType = statusVisitor.Visit(statusType, nil)
						//klog.Infof("Found swagger information for %v", typeName)
						newTypes.Add(astmodel.MakeTypeDefinition(typeName, resource.WithStatus(statusType)))
					} else {
						klog.Infof("No swagger information found for %v", typeName)
						newTypes.Add(typeDef)
					}
				} else {
					newTypes.Add(typeDef)
				}
			}

			klog.Infof("Input %v types, output %v types", len(types), len(newTypes))

			return newTypes, nil
		},
	}
}

func appendStatusToName(typeName astmodel.TypeName) astmodel.TypeName {
	return astmodel.MakeTypeName(typeName.PackageReference, typeName.Name()+"Status")
}

func makeStatusVisitor(statusTypes astmodel.Types) astmodel.TypeVisitor {
	visitor := astmodel.MakeTypeVisitor()

	// rename any references to object/resource types
	visitor.VisitTypeName = func(this *astmodel.TypeVisitor, it astmodel.TypeName, ctx interface{}) astmodel.Type {
		if statusType, ok := statusTypes[it]; ok {
			switch statusType.Type().(type) {
			case *astmodel.ResourceType:
			case *astmodel.ObjectType:
				return appendStatusToName(it)
			}
		}

		return it
	}

	return visitor
}

var swaggerVersionRegex = regexp.MustCompile("\\d{4}-\\d{2}-\\d{2}(-preview)?")
var swaggerGroupRegex = regexp.MustCompile("[Mm]icrosoft\\.[^/\\\\]+")

func loadSwaggerData(ctx context.Context, idFactory astmodel.IdentifierFactory, config *config.Configuration) (astmodel.Types, error) {

	azureRestSpecsRoot := "/home/george/k8s-infra/hack/generator/azure-rest-api-specs/specification"
	loaded, err := loadSchemas(ctx, azureRestSpecsRoot)
	if err != nil {
		return nil, err
	}

	statusTypes := make(astmodel.Types)

	extractor := typeExtractor{
		idFactory: idFactory,
		cache:     loaded.cache,
		config:    config,
	}

	for filePath, schema := range loaded.schemas {
		err := extractor.extractTypes(ctx, filePath, schema, statusTypes)
		if err != nil {
			return nil, err
		}
	}

	return statusTypes, nil
}

type loadResult struct {
	schemas map[string]spec.Swagger
	cache   *jsonast.OpenAPISchemaCache
}

func loadSchemas(ctx context.Context, rootPath string) (loadResult, error) {

	var wg sync.WaitGroup
	cache := jsonast.MakeOpenAPISchemaCache()

	var mutex sync.Mutex
	schemas := make(map[string]spec.Swagger)
	var sharedErr error

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if ctx.Err() != nil { // check for cancellation
			return ctx.Err()
		}

		// TODO: less hardcoding
		if filepath.Ext(path) == ".json" &&
			!strings.Contains(path, "/examples/") &&
			!strings.Contains(path, "/quickstart-templates/") &&
			!strings.Contains(path, "/control-plane/") &&
			!strings.Contains(path, "/data-plane/") {

			wg.Add(1)
			go func() {
				swagger, err := cache.PreloadCache(path)
				mutex.Lock()
				if err != nil {
					// first error wins
					if sharedErr == nil {
						sharedErr = err
					}
				} else {
					schemas[path] = swagger
				}
				mutex.Unlock()

				wg.Done()
			}()
		}

		return nil
	})

	wg.Wait()

	if err != nil {
		return loadResult{}, err
	}

	if sharedErr != nil {
		return loadResult{}, sharedErr
	}

	return loadResult{schemas, cache}, nil
}

type typeExtractor struct {
	idFactory astmodel.IdentifierFactory
	config    *config.Configuration
	cache     *jsonast.OpenAPISchemaCache
}

func (extractor *typeExtractor) extractTypes(
	ctx context.Context,
	filePath string,
	swagger spec.Swagger,
	types astmodel.Types) error {

	version := swagger.Info.Version
	packageName := extractor.idFactory.CreatePackageNameFromVersion(version)

	scanner := jsonast.NewSchemaScanner(extractor.idFactory, extractor.config)

	for operationPath, op := range swagger.Paths.Paths {
		put := op.Put
		if put == nil {
			continue
		}

		group := swaggerGroupRegex.FindString(operationPath)
		if group == "" {
			klog.Warningf("unable to extract group from %q", operationPath)
			continue
		}

		resourceName, err := extractor.resourceNameFromOperationPath(group, packageName, operationPath)
		if err != nil {
			klog.Errorf("unable to produce resource name for path %q (in %s)", operationPath, filePath)
			continue
		}

		resourceType, err := extractor.resourceTypeFromOperation(ctx, scanner, swagger, filePath, group, put)
		if err != nil {
			if err == context.Canceled {
				return err
			}

			klog.Errorf("unable to produce type for %v (in %s): %v", resourceName, filePath, err)
			continue // TODO: make fatal
		}

		if resourceType != nil {
			if existingResource, ok := types[resourceName]; !ok {
				types.Add(astmodel.MakeTypeDefinition(resourceName, resourceType))
			} else {
				if !astmodel.TypeEquals(existingResource.Type(), resourceType) {
					klog.Errorf("already defined differently: %v (in %s)", resourceName, filePath)
				}
			}
		}
	}

	for key, def := range scanner.Definitions() {
		// get generated-aside definitions too
		if _, ok := types[key]; !ok {
			types.Add(def)
		}
	}

	return nil
}

func (extractor *typeExtractor) resourceNameFromOperationPath(group string, packageName string, operationPath string) (astmodel.TypeName, error) {
	// infer name from path
	name, err := inferNameFromURLPath(group, operationPath)
	if err != nil {
		return astmodel.TypeName{}, errors.Wrapf(err, "unable to infer name from path")
	}

	packageRef := astmodel.MakeLocalPackageReference(extractor.idFactory.CreateGroupName(group), packageName)
	return astmodel.MakeTypeName(packageRef, name), nil
}

func (extractor *typeExtractor) resourceTypeFromOperation(
	ctx context.Context,
	scanner *jsonast.SchemaScanner,
	schemaRoot spec.Swagger,
	filePath string,
	group string,
	operation *spec.Operation) (astmodel.Type, error) {

	for _, param := range operation.Parameters {
		if param.In == "body" && param.Required { // assume this is the Resource
			schema := jsonast.MakeOpenAPISchema(*param.Schema, schemaRoot, filePath, group, extractor.cache)
			return scanner.RunHandlerForSchema(ctx, schema)
		}
	}

	return nil, nil
}

func inferNameFromURLPath(group string, operationPath string) (string, error) {
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
					return "", errors.Errorf("multiple parameters in path")
				}

				skippedLast = true
			}
		} else if strings.ToLower(urlPart) == strings.ToLower(group) {
			reading = true
		}
	}

	if reading == false {
		return "", errors.Errorf("no group name (‘Microsoft…’) found")
	}

	if name == "" {
		return "", errors.Errorf("couldn’t infer name")
	}

	return name, nil
}
