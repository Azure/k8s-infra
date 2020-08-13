/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package jsonast

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// GoJSONSchema implements the Schema abstraction for gojsonschema
type GoJSONSchema struct {
	inner *gojsonschema.SubSchema
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

func (schema GoJSONSchema) URL() *url.URL {
	return schema.inner.ID.GetUrl()
}

func (schema GoJSONSchema) Title() *string {
	return schema.inner.Title
}

func (schema GoJSONSchema) HasType(schemaType SchemaType) bool {
	return schema.inner.Types.Contains(string(schemaType))
}

func (schema GoJSONSchema) RequiredProperties() []string {
	return schema.inner.Required
}

func (schema GoJSONSchema) HasAllOf() bool {
	return len(schema.inner.AllOf) > 0
}

func (schema GoJSONSchema) AllOf() []Schema {
	return transformGoJSONSlice(schema.inner.AllOf)
}

func (schema GoJSONSchema) HasAnyOf() bool {
	return len(schema.inner.AnyOf) > 0
}

func (schema GoJSONSchema) AnyOf() []Schema {
	return transformGoJSONSlice(schema.inner.AnyOf)
}

func (schema GoJSONSchema) HasOneOf() bool {
	return len(schema.inner.OneOf) > 0
}

func (schema GoJSONSchema) OneOf() []Schema {
	return transformGoJSONSlice(schema.inner.OneOf)
}

func (schema GoJSONSchema) Properties() map[string]Schema {
	result := make(map[string]Schema)
	for _, prop := range schema.inner.PropertiesChildren {
		result[prop.Property] = GoJSONSchema{prop}
	}

	return result
}

func (schema GoJSONSchema) Description() *string {
	return schema.inner.Description
}

func (schema GoJSONSchema) Items() []Schema {
	return transformGoJSONSlice(schema.inner.ItemsChildren)
}

func (schema GoJSONSchema) AdditionalPropertiesAllowed() bool {
	aps := schema.inner.AdditionalProperties

	return aps == nil || aps != false
}

func (schema GoJSONSchema) AdditionalPropertiesSchema() Schema {
	result := schema.inner.AdditionalProperties

	if result == nil {
		return nil
	}

	return GoJSONSchema{result.(*gojsonschema.SubSchema)}
}

func (schema GoJSONSchema) EnumValues() []string {
	return schema.inner.Enum
}

func (schema GoJSONSchema) IsRef() bool {
	return schema.inner.RefSchema != nil
}

func (schema GoJSONSchema) RefSchema() Schema {
	return GoJSONSchema{schema.inner.RefSchema}
}

func isURLPathSeparator(c rune) bool {
	return c == '/'
}

func (schema GoJSONSchema) RefObjectName() (string, error) {
	url := schema.inner.Ref.GetUrl()
	fragmentParts := strings.FieldsFunc(url.Fragment, isURLPathSeparator)

	if len(fragmentParts) == 0 {
		panic(fmt.Sprintf("no fragment extracted from %s", url.String()))
	}

	return fragmentParts[len(fragmentParts)-1], nil
}

func (schema GoJSONSchema) RefGroupName() (string, error) {
	url := schema.inner.Ref.GetUrl()
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

var versionRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}(-preview)?`)

func (schema GoJSONSchema) RefVersion() (string, error) {
	url := schema.inner.Ref.GetUrl()
	pathParts := strings.FieldsFunc(url.Path, isURLPathSeparator)

	for _, p := range pathParts {
		if versionRegex.MatchString(p) {
			return p, nil
		}
	}

	// No version found, that's fine
	return "", nil
}

func (schema GoJSONSchema) RefIsResource() bool {
	url := schema.inner.Ref.GetUrl()
	fragmentParts := strings.FieldsFunc(url.Fragment, isURLPathSeparator)

	for _, fragmentPart := range fragmentParts {
		if fragmentPart == "resourceDefinitions" {
			return true
		}
	}

	return false
}
