/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package jsonast

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"k8s.io/klog/v2"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/devigned/tab"
	"github.com/xeipuuv/gojsonschema"
)

type (
	// SchemaType defines the type of JSON schema node we are currently processing
	SchemaType string

	// TypeHandler is a standard delegate used for walking the schema tree.
	// Note that it is permissible for a TypeHandler to return `nil, nil`, which indicates that
	// there is no type to be included in the output.
	TypeHandler func(ctx context.Context, scanner *SchemaScanner, schema *gojsonschema.SubSchema) (astmodel.Type, error)

	// UnknownSchemaError is used when we find a JSON schema node that we don't know how to handle
	UnknownSchemaError struct {
		Schema  *gojsonschema.SubSchema
		Filters []string
	}

	// A BuilderOption is used to provide custom configuration for our scanner
	BuilderOption func(scanner *SchemaScanner) error

	// A SchemaScanner is used to scan a JSON Schema extracting and collecting type definitions
	SchemaScanner struct {
		definitions  map[astmodel.TypeName]astmodel.TypeDefiner
		TypeHandlers map[SchemaType]TypeHandler
		Filters      []string
		idFactory    astmodel.IdentifierFactory
	}
)

// findTypeDefinition looks to see if we have seen the specified definition before, returning its definition if we have.
func (scanner *SchemaScanner) findTypeDefinition(name *astmodel.TypeName) (astmodel.TypeDefiner, bool) {
	result, ok := scanner.definitions[*name]
	return result, ok
}

// addTypeDefinition adds a type definition to emit later
func (scanner *SchemaScanner) addTypeDefinition(def astmodel.TypeDefiner) {
	scanner.definitions[*def.Name()] = def
}

// addEmptyTypeDefinition adds a placeholder definition; it should always be replaced later
func (scanner *SchemaScanner) addEmptyTypeDefinition(name *astmodel.TypeName) {
	scanner.definitions[*name] = nil
}

// removeTypeDefinition removes a type definition
func (scanner *SchemaScanner) removeTypeDefinition(name *astmodel.TypeName) {
	delete(scanner.definitions, *name)
}

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

	expressionFragment = "/definitions/expression"
)

func (use *UnknownSchemaError) Error() string {
	if use.Schema == nil || use.Schema.ID == nil {
		return fmt.Sprint("unable to determine schema type for nil schema or one without an ID")
	}
	return fmt.Sprintf("unable to determine the schema type for %s", use.Schema.ID.String())
}

// NewSchemaScanner constructs a new scanner, ready for use
func NewSchemaScanner(idFactory astmodel.IdentifierFactory) *SchemaScanner {
	return &SchemaScanner{
		definitions:  make(map[astmodel.TypeName]astmodel.TypeDefiner),
		TypeHandlers: DefaultTypeHandlers(),
		idFactory:    idFactory,
	}
}

// AddTypeHandler will override a default type handler for a given SchemaType. This allows for a consumer to customize
// AST generation.
func (scanner *SchemaScanner) AddTypeHandler(schemaType SchemaType, handler TypeHandler) {
	scanner.TypeHandlers[schemaType] = handler
}

// RunHandler triggers the appropriate handler for the specified schemaType
func (scanner *SchemaScanner) RunHandler(ctx context.Context, schemaType SchemaType, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	handler := scanner.TypeHandlers[schemaType]
	return handler(ctx, scanner, schema)
}

// RunHandlerForSchema inspects the passed schema to identify what kind it is, then runs the appropriate handler
func (scanner *SchemaScanner) RunHandlerForSchema(ctx context.Context, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	schemaType, err := getSubSchemaType(schema)
	if err != nil {
		return nil, err
	}

	return scanner.RunHandler(ctx, schemaType, schema)
}

// AddFilters will add a filter (perhaps not currently used?)
func (scanner *SchemaScanner) AddFilters(filters []string) {
	scanner.Filters = append(scanner.Filters, filters...)
}

// ToNodes takes in the resources section of the Azure deployment template schema and returns golang AST Packages
//    containing the types described in the schema which match the {resource_type}/{version} filters provided.
//
// 		The schema we are working with is something like the following (in yaml for brevity):
//
// 		resources:
// 			items:
// 				oneOf:
// 					allOf:
// 						$ref: {{ base resource schema for ARM }}
// 						oneOf:
// 							- ARM resources
// 				oneOf:
// 					allOf:
// 						$ref: {{ base resource for external resources, think SendGrid }}
// 						oneOf:
// 							- External ARM resources
// 				oneOf:
// 					allOf:
// 						$ref: {{ base resource for ARM specific stuff like locks, deployments, etc }}
// 						oneOf:
// 							- ARM specific resources. I'm not 100% sure why...
//
// 		allOf acts like composition which composites each schema from the child oneOf with the base reference from allOf.
func (scanner *SchemaScanner) GenerateDefinitions(ctx context.Context, schema *gojsonschema.SubSchema, opts ...BuilderOption) ([]astmodel.TypeDefiner, error) {
	ctx, span := tab.StartSpan(ctx, "ToNodes")
	defer span.End()

	for _, opt := range opts {
		if err := opt(scanner); err != nil {
			return nil, err
		}
	}

	// get initial topic from ID and Title:
	url := schema.ID.GetUrl()
	if schema.Title == nil {
		return nil, fmt.Errorf("Given schema has no Title")
	}

	rootName := *schema.Title

	rootGroup, err := groupOf(url)
	if err != nil {
		return nil, fmt.Errorf("Unable to extract group for schema: %w", err)
	}

	rootVersion, err := versionOf(url)
	if err != nil {
		return nil, fmt.Errorf("Unable to extract version for schema: %w", err)
	}

	rootPackage := astmodel.NewLocalPackageReference(
		scanner.idFactory.CreateGroupName(rootGroup),
		scanner.idFactory.CreatePackageNameFromVersion(rootVersion))

	rootTypeName := astmodel.NewTypeName(rootPackage, rootName)

	_, err = generateDefinitionsFor(ctx, scanner, rootTypeName, false, url, schema)
	if err != nil {
		return nil, err
	}

	// produce the results
	var defs []astmodel.TypeDefiner
	for _, def := range scanner.definitions {
		defs = append(defs, def)
	}

	return defs, nil
}

// DefaultTypeHandlers will create a default map of JSONType to AST transformers
func DefaultTypeHandlers() map[SchemaType]TypeHandler {
	return map[SchemaType]TypeHandler{
		Array:  arrayHandler,
		OneOf:  oneOfHandler,
		AnyOf:  anyOfHandler,
		AllOf:  allOfHandler,
		Ref:    refHandler,
		Object: objectHandler,
		Enum:   enumHandler,
		String: fixedTypeHandler(astmodel.StringType, "string"),
		Int:    fixedTypeHandler(astmodel.IntType, "int"),
		Number: fixedTypeHandler(astmodel.FloatType, "number"),
		Bool:   fixedTypeHandler(astmodel.BoolType, "bool"),
	}
}

func enumHandler(ctx context.Context, scanner *SchemaScanner, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	_, span := tab.StartSpan(ctx, "enumHandler")
	defer span.End()

	// Default to a string base type
	baseType := astmodel.StringType
	for _, t := range []SchemaType{Bool, Int, Number, String} {
		if schema.Types.Contains(string(t)) {
			bt, err := getPrimitiveType(t)
			if err != nil {
				return nil, err
			}

			baseType = bt
		}
	}

	var values []astmodel.EnumValue
	for _, v := range schema.Enum {
		id := scanner.idFactory.CreateIdentifier(v, astmodel.Public)
		values = append(values, astmodel.EnumValue{Identifier: id, Value: v})
	}

	enumType := astmodel.NewEnumType(baseType, values)

	return enumType, nil
}

func fixedTypeHandler(typeToReturn astmodel.Type, handlerName string) TypeHandler {
	return func(ctx context.Context, scanner *SchemaScanner, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
		_, span := tab.StartSpan(ctx, handlerName+"Handler")
		defer span.End()

		return typeToReturn, nil
	}
}

func objectHandler(ctx context.Context, scanner *SchemaScanner, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "objectHandler")
	defer span.End()

	fields, err := getFields(ctx, scanner, schema)
	if err != nil {
		return nil, err
	}

	// if we _only_ have an 'additionalProperties' field, then we are making
	// a dictionary-like type, and we won't generate a struct; instead, we
	// will just use the 'additionalProperties' type directly
	if len(fields) == 1 && fields[0].FieldName() == "additionalProperties" {
		return fields[0].FieldType(), nil
	}

	structDefinition := astmodel.NewStructType(fields...)
	return structDefinition, nil
}

func generateFieldDefinition(ctx context.Context, scanner *SchemaScanner, prop *gojsonschema.SubSchema) (*astmodel.FieldDefinition, error) {
	fieldName := scanner.idFactory.CreateFieldName(prop.Property, astmodel.Public)

	schemaType, err := getSubSchemaType(prop)
	if _, ok := err.(*UnknownSchemaError); ok {
		// if we don't know the type, we still need to provide the property, we will just provide open interface
		field := astmodel.NewFieldDefinition(fieldName, prop.Property, astmodel.AnyType)
		return field, nil
	}

	if err != nil {
		return nil, err
	}

	propType, err := scanner.RunHandler(ctx, schemaType, prop)
	if _, ok := err.(*UnknownSchemaError); ok {
		// if we don't know the type, we still need to provide the property, we will just provide open interface
		field := astmodel.NewFieldDefinition(fieldName, prop.Property, astmodel.AnyType)
		return field, nil
	}

	if err != nil {
		return nil, err
	}

	field := astmodel.NewFieldDefinition(fieldName, prop.Property, propType)
	return field, nil
}

func getFields(ctx context.Context, scanner *SchemaScanner, schema *gojsonschema.SubSchema) ([]*astmodel.FieldDefinition, error) {
	ctx, span := tab.StartSpan(ctx, "getFields")
	defer span.End()

	var fields []*astmodel.FieldDefinition
	for _, prop := range schema.PropertiesChildren {

		fieldDefinition, err := generateFieldDefinition(ctx, scanner, prop)
		if err != nil {
			return nil, err
		}

		// add documentation
		fieldDefinition = fieldDefinition.WithDescription(prop.Description)

		// add validations
		isRequired := false
		for _, required := range schema.Required {
			if prop.Property == required {
				isRequired = true
				break
			}
		}

		if isRequired {
			fieldDefinition = fieldDefinition.MakeRequired()
		} else {
			fieldDefinition = fieldDefinition.MakeOptional()
		}

		fields = append(fields, fieldDefinition)
	}

	// see: https://json-schema.org/understanding-json-schema/reference/object.html#properties
	if schema.AdditionalProperties == nil {
		// if not specified, any additional properties are allowed (TODO: tell all Azure teams this fact and get them to update their API definitions)
		// for now we aren't following the spec 100% as it pollutes the generated code
		// only generate this field if there are no other fields:
		if len(fields) == 0 {
			// TODO: for JSON serialization this needs to be unpacked into "parent"
			additionalPropsField := astmodel.NewFieldDefinition("additionalProperties", "additionalProperties", astmodel.NewStringMapType(astmodel.AnyType))
			fields = append(fields, additionalPropsField)
		}
	} else if schema.AdditionalProperties != false {
		// otherwise, if not false then it is a type for all additional fields
		// TODO: for JSON serialization this needs to be unpacked into "parent"
		additionalPropsType, err := scanner.RunHandlerForSchema(ctx, schema.AdditionalProperties.(*gojsonschema.SubSchema))
		if err != nil {
			return nil, err
		}

		additionalPropsField := astmodel.NewFieldDefinition(astmodel.FieldName("additionalProperties"), "additionalProperties", astmodel.NewStringMapType(additionalPropsType))
		fields = append(fields, additionalPropsField)
	}

	return fields, nil
}

func refHandler(ctx context.Context, scanner *SchemaScanner, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "refHandler")
	defer span.End()

	url := schema.Ref.GetUrl()

	if url.Fragment == expressionFragment {
		// skip expressions
		return nil, nil
	}

	// make a new topic based on the ref URL
	name, err := objectTypeOf(url)
	if err != nil {
		return nil, err
	}

	group, err := groupOf(url)
	if err != nil {
		return nil, err
	}

	version, err := versionOf(url)
	if err != nil {
		return nil, err
	}

	isResource := isResource(url)

	// produce a usable name:
	typeName := astmodel.NewTypeName(
		astmodel.NewLocalPackageReference(
			scanner.idFactory.CreateGroupName(group),
			scanner.idFactory.CreatePackageNameFromVersion(version)),
		scanner.idFactory.CreateIdentifier(name, astmodel.Public))

	return generateDefinitionsFor(ctx, scanner, typeName, isResource, url, schema.RefSchema)
}

func generateDefinitionsFor(ctx context.Context, scanner *SchemaScanner, typeName *astmodel.TypeName, isResource bool, url *url.URL, schema *gojsonschema.SubSchema) (astmodel.Type, error) {

	schemaType, err := getSubSchemaType(schema)
	if err != nil {
		return nil, err
	}

	// see if we already generated something for this ref
	if _, ok := scanner.findTypeDefinition(typeName); ok {
		return typeName, nil
	}

	// Add a placeholder to avoid recursive calls
	// we will overwrite this later
	scanner.addEmptyTypeDefinition(typeName)

	result, err := scanner.RunHandler(ctx, schemaType, schema)
	if err != nil {
		scanner.removeTypeDefinition(typeName) // we weren't able to generate it, remove placeholder
		return nil, err
	}

	// Give the type a name:
	definer, otherDefs := result.CreateDefinitions(typeName, scanner.idFactory, isResource)

	description := "Generated from: " + url.String()
	definer = definer.WithDescription(&description)

	// register all definitions
	scanner.addTypeDefinition(definer)
	for _, otherDef := range otherDefs {
		scanner.addTypeDefinition(otherDef)
	}

	// return the name of the primary type
	return definer.Name(), nil
}

func allOfHandler(ctx context.Context, scanner *SchemaScanner, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "allOfHandler")
	defer span.End()

	var fields []*astmodel.FieldDefinition
	for _, all := range schema.AllOf {

		d, err := scanner.RunHandlerForSchema(ctx, all)
		if err != nil {
			return nil, err
		}

		if d == nil {
			continue // ignore skipped types
		}

		// unpack the contents of what we got from subhandlers:
		switch s := d.(type) {
		case *astmodel.StructType:
			// if it's a struct type get all its fields:
			fields = append(fields, s.Fields()...)

		case *astmodel.TypeName:
			// TODO: need to check if this is a reference to a struct type or not

			// if it's a reference to a defined struct, embed it inside:
			fields = append(fields, astmodel.NewEmbeddedStructDefinition(s))

		default:
			klog.Errorf("Unhandled type in allOf: %#v\n", d)
		}
	}

	result := astmodel.NewStructType(fields...)
	return result, nil
}

func oneOfHandler(ctx context.Context, scanner *SchemaScanner, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "oneOfHandler")
	defer span.End()

	// make sure we visit everything before bailing out,
	// to get all types generated even if we can't use them
	var results []astmodel.Type
	for _, one := range schema.OneOf {
		result, err := scanner.RunHandlerForSchema(ctx, one)
		if err != nil {
			return nil, err
		}

		if result != nil {
			results = appendIfUniqueType(results, result)
		}
	}

	if len(results) == 1 {
		return results[0], nil
	}

	// If there's more than one option, synthesize a type.
	// Note that this is required because Kubernetes CRDs do not support OneOf the same way
	// OpenAPI does, see https://github.com/Azure/k8s-infra/issues/71
	var fields []*astmodel.FieldDefinition

	for i, t := range results {
		switch concreteType := t.(type) {
		case *astmodel.TypeName:
			// Just a sanity check that we've already scanned this definition
			// TODO: Could remove this?
			if _, ok := scanner.findTypeDefinition(concreteType); !ok {
				return nil, fmt.Errorf("couldn't find struct for definition: %v", concreteType)
			}
			fieldName := scanner.idFactory.CreateFieldName(concreteType.Name(), astmodel.Public)

			// JSON name is unimportant here because we will implement the JSON marshaller anyway,
			// but we still need it for controller-gen
			jsonName := scanner.idFactory.CreateIdentifier(concreteType.Name(), astmodel.Internal)
			field := astmodel.NewFieldDefinition(fieldName, jsonName, concreteType).MakeOptional()
			fields = append(fields, field)
		case *astmodel.EnumType:
			// TODO: This name sucks but what alternative do we have?
			name := fmt.Sprintf("enum%v", i)
			fieldName := scanner.idFactory.CreateFieldName(name, astmodel.Public)

			// JSON name is unimportant here because we will implement the JSON marshaller anyway,
			// but we still need it for controller-gen
			jsonName := scanner.idFactory.CreateIdentifier(name, astmodel.Internal)
			field := astmodel.NewFieldDefinition(fieldName, jsonName, concreteType).MakeOptional()
			fields = append(fields, field)
		case *astmodel.StructType:
			// TODO: This name sucks but what alternative do we have?
			name := fmt.Sprintf("object%v", i)
			fieldName := scanner.idFactory.CreateFieldName(name, astmodel.Public)

			// JSON name is unimportant here because we will implement the JSON marshaller anyway,
			// but we still need it for controller-gen
			jsonName := scanner.idFactory.CreateIdentifier(name, astmodel.Internal)
			field := astmodel.NewFieldDefinition(fieldName, jsonName, concreteType).MakeOptional()
			fields = append(fields, field)
		case *astmodel.PrimitiveType:
			var primitiveTypeName string
			if concreteType == astmodel.AnyType {
				primitiveTypeName = "anything"
			} else {
				primitiveTypeName = concreteType.Name()
			}

			// TODO: This name sucks but what alternative do we have?
			name := fmt.Sprintf("%v%v", primitiveTypeName, i)
			fieldName := scanner.idFactory.CreateFieldName(name, astmodel.Public)

			// JSON name is unimportant here because we will implement the JSON marshaller anyway,
			// but we still need it for controller-gen
			jsonName := scanner.idFactory.CreateIdentifier(name, astmodel.Internal)
			field := astmodel.NewFieldDefinition(fieldName, jsonName, concreteType).MakeOptional()
			fields = append(fields, field)
		default:
			return nil, fmt.Errorf("unexpected oneOf member, type: %T", t)
		}
	}

	structType := astmodel.NewStructType(fields...)
	structType = structType.WithFunction(
		"Marshal",
		astmodel.NewOneOfJSONMarshalFunction(structType, scanner.idFactory))

	return structType, nil
}

func anyOfHandler(ctx context.Context, scanner *SchemaScanner, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "anyOfHandler")
	defer span.End()

	// again, make sure we walk everything first
	// to generate types:
	var results []astmodel.Type
	for _, any := range schema.AnyOf {
		result, err := scanner.RunHandlerForSchema(ctx, any)
		if err != nil {
			return nil, err
		}

		if result != nil {
			results = append(results, result)
		}
	}

	if len(results) == 1 {
		return results[0], nil
	}

	// return all possibilities...
	return astmodel.AnyType, nil
}

func arrayHandler(ctx context.Context, scanner *SchemaScanner, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "arrayHandler")
	defer span.End()

	if len(schema.ItemsChildren) > 1 {
		return nil, fmt.Errorf("item contains more children than expected: %v", schema.ItemsChildren)
	}

	if len(schema.ItemsChildren) == 0 {
		// there is no type to the elements, so we must assume interface{}
		klog.Warning("Interface assumption unproven\n")

		return astmodel.NewArrayType(astmodel.AnyType), nil
	}

	// get the only child type and wrap it up as an array type:

	onlyChild := schema.ItemsChildren[0]

	astType, err := scanner.RunHandlerForSchema(ctx, onlyChild)
	if err != nil {
		return nil, err
	}

	return astmodel.NewArrayType(astType), nil
}

func getSubSchemaType(schema *gojsonschema.SubSchema) (SchemaType, error) {
	// handle special nodes:
	switch {
	case schema.Enum != nil: // this should come before the primitive checks below
		return Enum, nil
	case schema.OneOf != nil:
		return OneOf, nil
	case schema.AllOf != nil:
		return AllOf, nil
	case schema.AnyOf != nil:
		return AnyOf, nil
	case schema.RefSchema != nil:
		return Ref, nil
	}

	if schema.Types.IsTyped() {
		for _, t := range []SchemaType{Object, String, Number, Int, Bool, Array} {
			if schema.Types.Contains(string(t)) {
				return t, nil
			}
		}
	}

	// TODO: this whole switch is a bit wrong because type: 'object' can
	// be combined with OneOf/AnyOf/etc. still, it works okay for now...
	if !schema.Types.IsTyped() && schema.PropertiesChildren != nil {
		// no type but has properties, treat it as an object
		return Object, nil
	}

	return Unknown, &UnknownSchemaError{Schema: schema}
}

func getPrimitiveType(name SchemaType) (*astmodel.PrimitiveType, error) {
	switch name {
	case String:
		return astmodel.StringType, nil
	case Int:
		return astmodel.IntType, nil
	case Number:
		return astmodel.FloatType, nil
	case Bool:
		return astmodel.BoolType, nil
	default:
		return astmodel.AnyType, fmt.Errorf("%s is not a simple type and no ast.NewIdent can be created", name)
	}
}

func isURLPathSeparator(c rune) bool {
	return c == '/'
}

// Extract the name of an object from the supplied schema URL
func objectTypeOf(url *url.URL) (string, error) {
	fragmentParts := strings.FieldsFunc(url.Fragment, isURLPathSeparator)

	return fragmentParts[len(fragmentParts)-1], nil
}

// Extract the 'group' (here filename) of an object from the supplied schemaURL
func groupOf(url *url.URL) (string, error) {
	pathParts := strings.FieldsFunc(url.Path, isURLPathSeparator)

	file := pathParts[len(pathParts)-1]
	if !strings.HasSuffix(file, ".json") {
		return "", fmt.Errorf("Unexpected URL format (doesn't point to .json file)")
	}

	return strings.TrimSuffix(file, ".json"), nil
}

func isResource(url *url.URL) bool {
	fragmentParts := strings.FieldsFunc(url.Fragment, isURLPathSeparator)

	for _, fragmentPart := range fragmentParts {
		if fragmentPart == "resourceDefinitions" {
			return true
		}
	}

	return false
}

var versionRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

// Extract the name of an object from the supplied schema URL
func versionOf(url *url.URL) (string, error) {
	pathParts := strings.FieldsFunc(url.Path, isURLPathSeparator)

	for _, p := range pathParts {
		if versionRegex.MatchString(p) {
			return p, nil
		}
	}

	// No version found, that's fine
	return "", nil
}

func appendIfUniqueType(slice []astmodel.Type, item astmodel.Type) []astmodel.Type {
	found := false
	for _, r := range slice {
		if r.Equals(item) {
			found = true
			break
		}
	}

	if !found {
		slice = append(slice, item)
	}

	return slice
}
