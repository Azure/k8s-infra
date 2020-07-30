/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package jsonast

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"sync"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
)

// OpenAPISchema implements the Schema abstraction for go-openapi
type OpenAPISchema struct {
	schema    spec.Schema
	root      spec.Swagger
	fileName  string
	groupName string
	version   string

	cache *OpenAPISchemaCache
}

type OpenAPISchemaCache struct {
	mutex sync.RWMutex
	files map[string]spec.Swagger
}

func MakeOpenAPISchemaCache() *OpenAPISchemaCache {
	return &OpenAPISchemaCache{
		files: make(map[string]spec.Swagger),
	}
}

func MakeOpenAPISchema(
	schema spec.Schema,
	root spec.Swagger,
	fileName string,
	groupName string,
	version string,
	cache *OpenAPISchemaCache) Schema {
	return &OpenAPISchema{schema, root, fileName, groupName, version, cache}
}

func (it *OpenAPISchema) withNewSchema(newSchema spec.Schema) Schema {
	return &OpenAPISchema{
		newSchema,
		it.root,
		it.fileName,
		it.groupName,
		it.version,
		it.cache,
	}
}

var _ Schema = &OpenAPISchema{}

func (it *OpenAPISchema) transformOpenAPISlice(slice []spec.Schema) []Schema {
	result := make([]Schema, len(slice))
	for i := range slice {
		result[i] = it.withNewSchema(slice[i])
	}

	return result
}

func (it *OpenAPISchema) Title() *string {
	if len(it.schema.Title) == 0 {
		return nil // translate to optional
	}

	return &it.schema.Title
}

func (it *OpenAPISchema) URL() *url.URL {
	url, err := url.Parse(it.schema.ID)
	if err != nil {
		return nil
	}

	return url
}

func (it *OpenAPISchema) HasType(schemaType SchemaType) bool {
	return it.schema.Type.Contains(string(schemaType))
}

func (it *OpenAPISchema) HasAllOf() bool {
	return len(it.schema.AllOf) > 0
}

func (it *OpenAPISchema) AllOf() []Schema {
	return it.transformOpenAPISlice(it.schema.AllOf)
}

func (it *OpenAPISchema) HasAnyOf() bool {
	return len(it.schema.AnyOf) > 0
}

func (it *OpenAPISchema) AnyOf() []Schema {
	return it.transformOpenAPISlice(it.schema.AnyOf)
}

func (it *OpenAPISchema) HasOneOf() bool {
	return len(it.schema.OneOf) > 0
}

func (it *OpenAPISchema) OneOf() []Schema {
	return it.transformOpenAPISlice(it.schema.OneOf)
}

func (it *OpenAPISchema) RequiredProperties() []string {
	return it.schema.Required
}

func (it *OpenAPISchema) Properties() map[string]Schema {
	result := make(map[string]Schema)
	for propName, propSchema := range it.schema.Properties {
		result[propName] = it.withNewSchema(propSchema)
	}

	return result
}

func (it *OpenAPISchema) Description() *string {
	if len(it.schema.Description) == 0 {
		return nil
	}

	return &it.schema.Description
}

func (it *OpenAPISchema) Items() []Schema {
	if it.schema.Items.Schema != nil {
		return []Schema{it.withNewSchema(*it.schema.Items.Schema)}
	}

	return it.transformOpenAPISlice(it.schema.Items.Schemas)
}

func (it *OpenAPISchema) AdditionalPropertiesAllowed() bool {
	return it.schema.AdditionalProperties == nil || it.schema.AdditionalProperties.Allows
}

func (it *OpenAPISchema) AdditionalPropertiesSchema() Schema {
	if it.schema.AdditionalProperties == nil {
		return nil
	}

	result := it.schema.AdditionalProperties.Schema
	if result == nil {
		return nil
	}

	return it.withNewSchema(*result)
}

func (it *OpenAPISchema) EnumValues() []string {
	result := make([]string, len(it.schema.Enum))
	for i, enumValue := range it.schema.Enum {
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

func (it *OpenAPISchema) IsRef() bool {
	return it.schema.Ref.GetURL() != nil
}

type fileNameAndSwagger struct {
	fileName string
	swagger  spec.Swagger
}

// loadFileWithoutCache loads the schema at the specified file path. it does not read from
// or add to the cache.
func (fileCache *OpenAPISchemaCache) loadUncachedFile(filePath string) (spec.Swagger, error) {
	var swagger spec.Swagger

	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return swagger, errors.Wrap(err, "unable to read swagger file")
	}

	err = swagger.UnmarshalJSON(fileContent)
	if err != nil {
		return swagger, errors.Wrap(err, "unable to parse swagger file")
	}

	return swagger, err
}

// PreloadCache loads the specified schema into the cache. it does not attempt to deduplicate work.
func (fileCache *OpenAPISchemaCache) PreloadCache(filePath string) (spec.Swagger, error) {
	swagger, err := fileCache.loadUncachedFile(filePath)
	if err == nil {
		fileCache.setCachedValue(filePath, swagger)
	}

	return swagger, err
}

func (fileCache *OpenAPISchemaCache) setCachedValue(filePath string, swagger spec.Swagger) {
	fileCache.mutex.Lock()
	defer fileCache.mutex.Unlock()
	fileCache.files[filePath] = swagger
}

// fetchFileRelative fetches the schema for the relative path created by combining 'baseFileName' and 'url'
// if multiple requests for the same file come in at the same time, only one request will hit the disk
func (fileCache *OpenAPISchemaCache) fetchFileRelative(baseFileName string, url *url.URL) (fileNameAndSwagger, error) {
	if url.IsAbs() {
		panic("only relative URLs may be passed")
	}

	fileURL, err := url.Parse("file://" + baseFileName)
	if err != nil {
		panic(err)
	}

	resolvedFile := fileURL.ResolveReference(url).Path
	swagger, err := fileCache.fetchFileAbsolute(resolvedFile)
	if err != nil {
		return fileNameAndSwagger{}, err
	}

	return fileNameAndSwagger{resolvedFile, swagger}, nil
}

// fetchFileAbsolute fetches the schema for the absolute path specified
// if multiple requests for the same file come in at the same time, only one request will hit the disk
func (fileCache *OpenAPISchemaCache) fetchFileAbsolute(filePath string) (spec.Swagger, error) {
	{
		fileCache.mutex.RLock()

		if swag, ok := fileCache.files[filePath]; ok {
			fileCache.mutex.RUnlock()
			return swag, nil
		}

		fileCache.mutex.RUnlock()
	}

	fileCache.mutex.Lock()
	defer fileCache.mutex.Unlock()

	// need to double-check after releasing read lock and claiming write lock
	if swag, ok := fileCache.files[filePath]; ok {
		return swag, nil
	}

	// ok, it really doesn't exist: read it
	swagger, err := fileCache.loadUncachedFile(filePath)

	if err == nil {
		fileCache.files[filePath] = swagger
	}

	return swagger, err
}

func (it *OpenAPISchema) RefSchema() Schema {
	var fileName string
	var root spec.Swagger
	if !it.schema.Ref.HasFragmentOnly {
		loaded, err := it.cache.fetchFileRelative(it.fileName, it.schema.Ref.GetURL())
		if err != nil {
			panic(err)
		}

		root = loaded.swagger
		fileName = loaded.fileName
	} else {
		root = it.root
		fileName = it.fileName
	}

	reffed := objectNameFromPointer(it.schema.Ref.GetPointer())
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
			// issues with types from the 'common-types' files which have no group and a version of 'v1'.
			it.groupName,
			it.version,
			it.cache,
		}
	}
}

func (it *OpenAPISchema) RefVersion() (string, error) {
	return it.version, nil
}

func (it *OpenAPISchema) RefGroupName() (string, error) {
	return it.groupName, nil
}

func (it *OpenAPISchema) RefObjectName() (string, error) {
	return objectNameFromPointer(it.schema.Ref.GetPointer()), nil
}

func objectNameFromPointer(ptr *jsonpointer.Pointer) string {
	// turns a fragment like "#/definitions/Name" into "Name"
	tokens := ptr.DecodedTokens()
	if len(tokens) != 2 || tokens[0] != "definitions" {
		// this condition is never violated by the swagger files
		panic(fmt.Sprintf("not understood: %v", tokens))
	}

	return tokens[1]
}

func (it *OpenAPISchema) RefIsResource() bool {
	return false
}
