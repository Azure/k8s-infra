/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"io/ioutil"
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

			statusLookup, err := loadSwaggerData(ctx, idFactory, config)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to load Swagger data")
			}

			klog.Infof("loaded Swagger data (%v types)", len(statusLookup))

			newTypes := make(astmodel.Types)

			statusVisitor := makeStatusVisitor(statusLookup)

			for typeName, typeDef := range types {
				// insert all status objects regardless of if they are used;
				// they will be trimmed by a later phase if they are not
				if statusDef, ok := statusLookup[typeName]; ok {
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
					if statusDef, ok := statusLookup[typeName]; ok {
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

	loader := fileLoader{config, idFactory, jsonast.MakeOpenAPISchemaCache()}
	output := makeOutput()

	err := filepath.Walk(azureRestSpecsRoot, func(path string, info os.FileInfo, err error) error {
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

			/*
				version := swaggerVersionRegex.FindString(path)
				if version == "" {
					//return fmt.Errorf("unable to extract version from path [%s]", path)
					return nil
				}
			*/

			/*
				group := swaggerGroupRegex.FindString(path)
				if group == "" {
					//return fmt.Errorf("unable to extract group from path [%s]", path)
					return nil
				}
			*/

			output.WaitGroup.Add(1)
			go loader.loadFileAsync(ctx, path, output)
		}

		return nil
	})

	output.WaitGroup.Wait()

	if err != nil {
		return nil, err
	}

	if output.err != nil {
		return nil, output.err
	}

	return output.results, nil
}

type loadOutput struct {
	sync.Mutex
	results astmodel.Types
	err     error
	sync.WaitGroup
}

func makeOutput() *loadOutput {
	return &loadOutput{
		results: make(astmodel.Types),
	}
}

func (output *loadOutput) setError(err error) {
	output.Mutex.Lock()
	defer output.Mutex.Unlock()

	// first one wins
	if output.err == nil {
		output.err = err
	}
}

func (output *loadOutput) setResults(results astmodel.Types) {
	output.Mutex.Lock()
	defer output.Mutex.Unlock()

	for key, value := range results {
		output.results[key] = value
	}
}

type fileLoader struct {
	config    *config.Configuration
	idFactory astmodel.IdentifierFactory
	cache     *jsonast.OpenAPISchemaCache
}

func (loader *fileLoader) loadFileAsync(
	ctx context.Context,
	path string,
	output *loadOutput) {

	results, err := loader.loadFile(ctx, path)
	if err != nil {
		output.setError(err)
	} else {
		output.setResults(results)
	}

	output.WaitGroup.Done()
}

func (loader *fileLoader) loadFile(
	ctx context.Context,
	filePath string) (astmodel.Types, error) {

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open Swagger document")
	}

	var swagger spec.Swagger
	err = swagger.UnmarshalJSON(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse Swagger document")
	}

	version := swagger.Info.Version
	packageName := loader.idFactory.CreatePackageNameFromVersion(version)

	scanner := jsonast.NewSchemaScanner(loader.idFactory, loader.config)

	results := make(astmodel.Types)
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

		groupName := loader.idFactory.CreateGroupName(group)

		// infer name from path
		name, err := inferNameFromURLPath(group, operationPath)
		if err != nil {
			klog.Warningf("unable to infer name from path: %s (in %s)", err.Error(), filePath)
			continue
		}

		packageRef := astmodel.MakeLocalPackageReference(groupName, packageName)
		typeName := astmodel.MakeTypeName(packageRef, name)
		//fmt.Printf("inferred name %v\n", typeName)

		for _, param := range put.Parameters {
			if param.In == "body" && param.Required { // assume this is the Resource
				schema := jsonast.MakeOpenAPISchema(*param.Schema, swagger, filePath, group, version, loader.cache)
				resultType, err := scanner.RunHandlerForSchema(ctx, schema)
				if err == nil {
					if definedType, ok := results[typeName]; !ok {
						results.Add(astmodel.MakeTypeDefinition(typeName, resultType))
					} else {
						if !astmodel.TypeEquals(definedType.Type(), resultType) {
							klog.Errorf("already defined differently: %v (in %s)", typeName, filePath)
						}
					}
				} else {
					if err == context.Canceled {
						return nil, err
					}

					klog.Errorf("unable to produce type for %s: %v", name, err)
					continue
					//return nil, err
				}

				break
			}
		}
	}

	for key, def := range scanner.Definitions() {
		// get generated-aside definitions too
		if _, ok := results[key]; !ok {
			results.Add(def)
		}
	}

	return results, nil
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
					return "", errors.Errorf("don’t know how to handle path %q", operationPath)
				}

				skippedLast = true
			}
		} else if strings.ToLower(urlPart) == strings.ToLower(group) {
			reading = true
		}
	}

	if reading == false {
		return "", errors.Errorf("no group name (‘Microsoft…’) found in path %q", operationPath)
	}

	if name == "" {
		return "", errors.Errorf("couldn’t get name from %q", operationPath)
	}

	return name, nil
}
