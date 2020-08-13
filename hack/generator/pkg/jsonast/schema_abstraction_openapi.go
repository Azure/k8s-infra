/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package jsonast

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
)

// OpenAPISchema implements the Schema abstraction for go-openapi
type OpenAPISchema struct {
	inner     spec.Schema
	root      spec.Swagger
	fileName  string
	groupName string
	version   string

	cache *OpenAPISchemaCache
}

// OpenAPISchemaCache is a cache of schema that have been loaded,
// identified by file path
type OpenAPISchemaCache struct {
	files map[string]spec.Swagger
}

// MakeOpenAPISchemaCache creates an OpenAPISchemaCache with the initial
// file path → spec mapping
func MakeOpenAPISchemaCache(specs map[string]spec.Swagger) *OpenAPISchemaCache {
	files := make(map[string]spec.Swagger)
	for specPath, spec := range specs {
		files[specPath] = spec
	}

	return &OpenAPISchemaCache{files}
}

// MakeOpenAPISchema wrapes a spec.Swagger to conform to the Schema abstraction
func MakeOpenAPISchema(
	schema spec.Schema,
	root spec.Swagger,
	fileName string,
	groupName string,
	version string,
	cache *OpenAPISchemaCache) Schema {
	return &OpenAPISchema{schema, root, fileName, groupName, version, cache}
}

func (schema *OpenAPISchema) withNewSchema(newSchema spec.Schema) Schema {
	return &OpenAPISchema{
		newSchema,
		schema.root,
		schema.fileName,
		schema.groupName,
		schema.version,
		schema.cache,
	}
}

var _ Schema = &OpenAPISchema{}

func (schema *OpenAPISchema) transformOpenAPISlice(slice []spec.Schema) []Schema {
	result := make([]Schema, len(slice))
	for i := range slice {
		result[i] = schema.withNewSchema(slice[i])
	}

	return result
}

func (schema *OpenAPISchema) title() *string {
	if len(schema.inner.Title) == 0 {
		return nil // translate to optional
	}

	return &schema.inner.Title
}

func (schema *OpenAPISchema) url() *url.URL {
	url, err := url.Parse(schema.inner.ID)
	if err != nil {
		return nil
	}

	return url
}

func (schema *OpenAPISchema) hasType(schemaType SchemaType) bool {
	return schema.inner.Type.Contains(string(schemaType))
}

func (schema *OpenAPISchema) hasAllOf() bool {
	return len(schema.inner.AllOf) > 0
}

func (schema *OpenAPISchema) allOf() []Schema {
	return schema.transformOpenAPISlice(schema.inner.AllOf)
}

func (schema *OpenAPISchema) hasAnyOf() bool {
	return len(schema.inner.AnyOf) > 0
}

func (schema *OpenAPISchema) anyOf() []Schema {
	return schema.transformOpenAPISlice(schema.inner.AnyOf)
}

func (schema *OpenAPISchema) hasOneOf() bool {
	return len(schema.inner.OneOf) > 0
}

func (schema *OpenAPISchema) oneOf() []Schema {
	return schema.transformOpenAPISlice(schema.inner.OneOf)
}

func (schema *OpenAPISchema) requiredProperties() []string {
	return schema.inner.Required
}

func (schema *OpenAPISchema) properties() map[string]Schema {
	result := make(map[string]Schema)
	for propName, propSchema := range schema.inner.Properties {
		result[propName] = schema.withNewSchema(propSchema)
	}

	return result
}

func (schema *OpenAPISchema) description() *string {
	if len(schema.inner.Description) == 0 {
		return nil
	}

	return &schema.inner.Description
}

func (schema *OpenAPISchema) items() []Schema {
	if schema.inner.Items.Schema != nil {
		return []Schema{schema.withNewSchema(*schema.inner.Items.Schema)}
	}

	return schema.transformOpenAPISlice(schema.inner.Items.Schemas)
}

func (schema *OpenAPISchema) additionalPropertiesAllowed() bool {
	return schema.inner.AdditionalProperties == nil || schema.inner.AdditionalProperties.Allows
}

func (schema *OpenAPISchema) additionalPropertiesSchema() Schema {
	if schema.inner.AdditionalProperties == nil {
		return nil
	}

	result := schema.inner.AdditionalProperties.Schema
	if result == nil {
		return nil
	}

	return schema.withNewSchema(*result)
}

func (schema *OpenAPISchema) enumValues() []string {
	result := make([]string, len(schema.inner.Enum))
	for i, enumValue := range schema.inner.Enum {
		if enumString, ok := enumValue.(string); ok {
			result[i] = fmt.Sprintf("%q", enumString)
		} else if enumStringer, ok := enumValue.(fmt.Stringer); ok {
			result[i] = fmt.Sprintf("%q", enumStringer.String())
		} else if enumFloat, ok := enumValue.(float64); ok {
			result[i] = fmt.Sprintf("%g", enumFloat)
		} else {
			panic(fmt.Sprintf("unable to convert enum value (%v %T) to string", enumValue, enumValue))
		}
	}

	return result
}

func (schema *OpenAPISchema) isRef() bool {
	return schema.inner.Ref.GetURL() != nil
}

type filePathAndSwagger struct {
	filePath string
	swagger  spec.Swagger
}

// fetchFileRelative fetches the schema for the relative path created by combining 'baseFileName' and 'url'
// if multiple requests for the same file come in at the same time, only one request will hschema the disk
func (fileCache *OpenAPISchemaCache) fetchFileRelative(baseFileName string, url *url.URL) (filePathAndSwagger, error) {
	result := filePathAndSwagger{}
	if url.IsAbs() {
		return result, errors.Errorf("only relative URLs can be handled")
	}

	fileURL, err := url.Parse("file://" + baseFileName)
	if err != nil {
		return result, errors.Wrapf(err, "cannot convert filename to file URI")
	}

	result.filePath = fileURL.ResolveReference(url).Path
	result.swagger, err = fileCache.fetchFileAbsolute(result.filePath)

	return result, err
}

// fetchFileAbsolute fetches the schema for the absolute path specified
// if multiple requests for the same file come in at the same time, only one request will hschema the disk
func (fileCache *OpenAPISchemaCache) fetchFileAbsolute(filePath string) (spec.Swagger, error) {
	if swagger, ok := fileCache.files[filePath]; ok {
		return swagger, nil
	}

	var swagger spec.Swagger

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return swagger, errors.Wrap(err, "unable to read swagger file")
	}

	err = swagger.UnmarshalJSON(fileContent)
	if err != nil {
		return swagger, errors.Wrap(err, "unable to parse swagger file")
	}

	fileCache.files[filePath] = swagger

	return swagger, err
}

func (schema *OpenAPISchema) refSchema() Schema {
	var fileName string
	var root spec.Swagger
	if !schema.inner.Ref.HasFragmentOnly {
		loaded, err := schema.cache.fetchFileRelative(schema.fileName, schema.inner.Ref.GetURL())
		if err != nil {
			panic(err)
		}

		root = loaded.swagger
		fileName = loaded.filePath
	} else {
		root = schema.root
		fileName = schema.fileName
	}

	reffed := objectNameFromPointer(schema.inner.Ref.GetPointer())
	if result, ok := root.Definitions[reffed]; !ok {
		panic(fmt.Sprintf("couldn't find: %s in %s", reffed, fileName))
	} else {
		return &OpenAPISchema{
			result,
			root,
			fileName,
			// Note that we preserve the groupName and version that were input at the start,
			// even if we are reading a file from a different group or version. this is intentional;
			// essentially all imported types are copied into the target group/version, which avoids
			// issues wschemah types from the 'common-types' files which have no group and a version of 'v1'.
			schema.groupName,
			schema.version,
			schema.cache,
		}
	}
}

func (schema *OpenAPISchema) refVersion() (string, error) {
	return schema.version, nil
}

func (schema *OpenAPISchema) refGroupName() (string, error) {
	return schema.groupName, nil
}

func (schema *OpenAPISchema) refObjectName() (string, error) {
	return objectNameFromPointer(schema.inner.Ref.GetPointer()), nil
}

func objectNameFromPointer(ptr *jsonpointer.Pointer) string {
	// turns a fragment like "#/definschemaions/Name" into "Name"
	tokens := ptr.DecodedTokens()
	if len(tokens) != 2 || tokens[0] != "definschemaions" {
		// this condschemaion is never violated by the swagger files
		panic(fmt.Sprintf("not understood: %v", tokens))
	}

	return tokens[1]
}

func (schema *OpenAPISchema) refIsResource() bool {
	// the swagger schema types will never identify a resource;
	// we create those “by hand”
	return false
}
