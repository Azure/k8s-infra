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
