/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package jsonast

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// SchemaType defines the type of JSON schema node we are currently processing
type SchemaType string

// Definitions for different kinds of JSON schema
const (
	AnyOf   SchemaType = "anyOf"
	AllOf   SchemaType = "allOf"
	OneOf   SchemaType = "oneOf"
	Ref     SchemaType = "ref"
	Array   SchemaType = "array"
	Bool    SchemaType = "boolean"
	Int     SchemaType = "integer"
	Number  SchemaType = "number"
	Object  SchemaType = "object"
	String  SchemaType = "string"
	Enum    SchemaType = "enum"
	Unknown SchemaType = "unknown"
)

type Schema interface {
	URL() *url.URL
	Title() *string
	Description() *string

	HasType(schemaType SchemaType) bool

	// complex things
	HasOneOf() bool
	OneOf() []Schema

	HasAnyOf() bool
	AnyOf() []Schema

	HasAllOf() bool
	AllOf() []Schema

	// enum
	EnumValues() []string

	// array things
	Items() []Schema

	// object things
	RequiredProperties() []string
	Properties() map[string]Schema
	AdditionalPropertiesAllowed() bool
	AdditionalPropertiesSchema() Schema

	// ref things
	IsRef() bool
	RefIsResource() bool
	RefGroupName() (string, error)
	RefObjectName() (string, error)
	RefVersion() (string, error)
	RefSchema() Schema
}

// GoJSONSchema implements the Schema abstraction for gojsonschema
type GoJSONSchema struct {
	schema *gojsonschema.SubSchema
}

func MakeGoJSONSchema(schema *gojsonschema.SubSchema) Schema {
	return GoJSONSchema{schema}
}

var _ Schema = GoJSONSchema{}

func transformGoJSONSlice(slice []*gojsonschema.SubSchema) []Schema {
	result := make([]Schema, len(slice))
	for i := range slice {
		result[i] = GoJSONSchema{slice[i]}
	}

	return result
}

func (it GoJSONSchema) URL() *url.URL {
	return it.schema.ID.GetUrl()
}

func (it GoJSONSchema) Title() *string {
	return it.schema.Title
}

func (it GoJSONSchema) HasType(schemaType SchemaType) bool {
	return it.schema.Types.Contains(string(schemaType))
}

func (it GoJSONSchema) RequiredProperties() []string {
	return it.schema.Required
}

func (it GoJSONSchema) HasAllOf() bool {
	return len(it.schema.AllOf) > 0
}

func (it GoJSONSchema) AllOf() []Schema {
	return transformGoJSONSlice(it.schema.AllOf)
}

func (it GoJSONSchema) HasAnyOf() bool {
	return len(it.schema.AnyOf) > 0
}

func (it GoJSONSchema) AnyOf() []Schema {
	return transformGoJSONSlice(it.schema.AnyOf)
}

func (it GoJSONSchema) HasOneOf() bool {
	return len(it.schema.OneOf) > 0
}

func (it GoJSONSchema) OneOf() []Schema {
	return transformGoJSONSlice(it.schema.OneOf)
}

func (it GoJSONSchema) Properties() map[string]Schema {
	result := make(map[string]Schema)
	for _, prop := range it.schema.PropertiesChildren {
		result[prop.Property] = GoJSONSchema{prop}
	}

	return result
}

func (it GoJSONSchema) Description() *string {
	return it.schema.Description
}

func (it GoJSONSchema) Items() []Schema {
	return transformGoJSONSlice(it.schema.ItemsChildren)
}

func (it GoJSONSchema) AdditionalPropertiesAllowed() bool {
	aps := it.schema.AdditionalProperties

	return aps == nil || aps != false
}

func (it GoJSONSchema) AdditionalPropertiesSchema() Schema {
	result := it.schema.AdditionalProperties

	if result == nil {
		return nil
	}

	return GoJSONSchema{result.(*gojsonschema.SubSchema)}
}

func (it GoJSONSchema) EnumValues() []string {
	return it.schema.Enum
}

func (it GoJSONSchema) IsRef() bool {
	return it.schema.RefSchema != nil
}

func (it GoJSONSchema) RefSchema() Schema {
	return GoJSONSchema{it.schema.RefSchema}
}

func isURLPathSeparator(c rune) bool {
	return c == '/'
}

func (it GoJSONSchema) RefObjectName() (string, error) {
	url := it.schema.Ref.GetUrl()
	fragmentParts := strings.FieldsFunc(url.Fragment, isURLPathSeparator)

	if len(fragmentParts) == 0 {
		panic(fmt.Sprintf("no fragment extracted from %s", url.String()))
	}

	return fragmentParts[len(fragmentParts)-1], nil
}

func (it GoJSONSchema) RefGroupName() (string, error) {
	url := it.schema.Ref.GetUrl()
	pathParts := strings.FieldsFunc(url.Path, isURLPathSeparator)

	if len(pathParts) == 0 {
		panic(fmt.Sprintf("no fields extracted from %s", url.String()))
	}

	file := pathParts[len(pathParts)-1]
	if !strings.HasSuffix(file, ".json") {
		return "", errors.Errorf("Unexpected URL format (doesn't point to .json file)")
	}

	return strings.TrimSuffix(file, ".json"), nil
}

var versionRegex = regexp.MustCompile(`v1|(\d{4}-\d{2}-\d{2}(-preview)?)`)

func (it GoJSONSchema) RefVersion() (string, error) {
	url := it.schema.Ref.GetUrl()
	pathParts := strings.FieldsFunc(url.Path, isURLPathSeparator)

	for _, p := range pathParts {
		if versionRegex.MatchString(p) {
			return p, nil
		}
	}

	// No version found, that's fine
	return "", nil
}

func (it GoJSONSchema) RefIsResource() bool {
	url := it.schema.Ref.GetUrl()
	fragmentParts := strings.FieldsFunc(url.Fragment, isURLPathSeparator)

	for _, fragmentPart := range fragmentParts {
		if fragmentPart == "resourceDefinitions" {
			return true
		}
	}

	return false
}

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

func objectNameFromPointer(ptr *jsonpointer.Pointer) string {
	tokens := ptr.DecodedTokens()
	if len(tokens) != 2 || tokens[0] != "definitions" {
		panic(fmt.Sprintf("not understood: %v", tokens))
	}

	return tokens[1]
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
			// TODO: dubious, should be based on fileName
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

func (it *OpenAPISchema) RefIsResource() bool {
	return false
}
