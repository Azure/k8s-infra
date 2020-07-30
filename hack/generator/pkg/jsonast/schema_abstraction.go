/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package jsonast

import (
	"net/url"
)

// Schema abstracts over the exact implementation of
// JSON schema we are using as we need to be able to
// process inputs provided by both 'gojsonschema' and
// 'go-openapi'. It is possible we could switch to
// 'go-openapi' for everything because it could handle
// both JSON Schema and Swagger; but this is not yet done.
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

	// enum things
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
