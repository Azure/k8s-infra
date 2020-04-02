package jsonast

import (
	"context"
	"fmt"
	"go/ast"
	"net/url"
	"strings"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/devigned/tab"
	"github.com/xeipuuv/gojsonschema"
)

type (
	SchemaType string

	// TypeHandler is a standard delegate used for walking the schema tree
	TypeHandler func(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error)

	UnknownSchemaError struct {
		Schema  *gojsonschema.SubSchema
		Filters []string
	}

	BuilderOption func(scanner *SchemaScanner) error

	// A SchemaScanner is used to scan a JSON Schema extracting and collecting type definitions
	SchemaScanner struct {
		Structs      []*astmodel.StructDefinition
		TypeHandlers map[SchemaType]TypeHandler
		Filters      []string
	}
)

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
	None    SchemaType = "none"
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
func NewSchemaScanner() *SchemaScanner {
	scanner := &SchemaScanner{}
	scanner.TypeHandlers = scanner.DefaultTypeHandlers()
	return scanner
}

// AddTypeHandler will override a default type handler for a given SchemaType. This allows for a consumer to customize
// AST generation.
func (scanner *SchemaScanner) AddTypeHandler(schemaType SchemaType, handler TypeHandler) {
	scanner.TypeHandlers[schemaType] = handler
}

// AddFilters will add a filter (perhaps not currently used?)
func (scanner *SchemaScanner) AddFilters(filters []string) {
	scanner.Filters = append(scanner.Filters, filters...)
}

/* ToNodes takes in the resources section of the Azure deployment template schema and returns golang AST Packages
   containing the types described in the schema which match the {resource_type}/{version} filters provided.

		The schema we are working with is something like the following (in yaml for brevity):

		resources:
			items:
				oneOf:
					allOf:
						$ref: {{ base resource schema for ARM }}
						oneOf:
							- ARM resources
				oneOf:
					allOf:
						$ref: {{ base resource for external resources, think SendGrid }}
						oneOf:
							- External ARM resources
				oneOf:
					allOf:
						$ref: {{ base resource for ARM specific stuff like locks, deployments, etc }}
						oneOf:
							- ARM specific resources. I'm not 100% sure why...

		allOf acts like composition which composites each schema from the child oneOf with the base reference from allOf.
*/
func (scanner *SchemaScanner) ToNodes(ctx context.Context, schema *gojsonschema.SubSchema, opts ...BuilderOption) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "ToNodes")
	defer span.End()

	for _, opt := range opts {
		if err := opt(scanner); err != nil {
			return nil, err
		}
	}

	schemaType, err := getSubSchemaType(schema)
	if err != nil {
		return nil, err
	}

	//TODO: Is empty string the right default here?
	topic := NewObjectScannerTopic(schema.Property, "")

	rootHandler := scanner.TypeHandlers[schemaType]
	nodes, err := rootHandler(ctx, topic, schema)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// DefaultTypeHandlers will create a default map of JSONType to AST transformers
func (scanner *SchemaScanner) DefaultTypeHandlers() map[SchemaType]TypeHandler {
	return map[SchemaType]TypeHandler{
		Array:  scanner.arrayHandler,
		OneOf:  scanner.oneOfHandler,
		AnyOf:  scanner.anyOfHandler,
		AllOf:  scanner.allOfHandler,
		Ref:    scanner.refHandler,
		Object: scanner.objectHandler,
		String: scanner.stringHandler,
		Int:    scanner.intHandler,
		Number: scanner.numberHandler,
		Bool:   scanner.boolHandler,
		Enum:   scanner.enumHandler,
		None:   scanner.noneHandler,
	}
}

func (scanner *SchemaScanner) enumHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "enumHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, String)
	if err != nil {
		return nil, err
	}

	return []astmodel.Definition{
		field,
	}, nil
}

func (scanner *SchemaScanner) noneHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "noneHandler")
	defer span.End()

	fmt.Printf("AST162 STA noneHandler")
	defer fmt.Printf("AST163 FIN noneHandler\n")

	fields, err := scanner.getFields(ctx, topic, schema)
	if err != nil {
		return nil, err
	}

	result := make([]astmodel.Definition, len(fields))
	for i := range fields {
		f := *fields[i]
		result[i] = f
	}

	return result, nil
}

func (scanner *SchemaScanner) boolHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "boolHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, Bool)
	if err != nil {
		return nil, err
	}

	return []astmodel.Definition{
		field,
	}, nil
}

func (scanner *SchemaScanner) numberHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "numberHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, Number)
	if err != nil {
		return nil, err
	}

	return []astmodel.Definition{
		field,
	}, nil
}

func (scanner *SchemaScanner) intHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "intHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, Int)
	if err != nil {
		return nil, err
	}

	return []astmodel.Definition{
		field,
	}, nil
}

func (scanner *SchemaScanner) stringHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "stringHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, String)
	if err != nil {
		return nil, err
	}

	return []astmodel.Definition{
		field,
	}, nil
}

func newField(ctx context.Context, fieldName string, fieldType string, description *string) (*astmodel.FieldDefinition, error) {
	ctx, span := tab.StartSpan(ctx, "newField")
	defer span.End()

	result := *astmodel.NewFieldDefinition(fieldName, fieldType)
	if description != nil {
		result = result.WithDescription(*description)
	}

	return &result, nil
}

func newPrimitiveField(ctx context.Context, schema *gojsonschema.SubSchema, typeIdent SchemaType) (*astmodel.FieldDefinition, error) {
	ctx, span := tab.StartSpan(ctx, "newPrimitiveField")
	defer span.End()

	ident, err := getPrimitiveType(typeIdent)
	if err != nil {
		return nil, err
	}

	result := *astmodel.NewFieldDefinition(schema.Property, ident)
	if schema.Description != nil {
		result = result.WithDescription(*schema.Description)
	}

	return &result, nil
}

func newStructField(ctx context.Context, schema *gojsonschema.SubSchema, structType *ast.StructType) (*astmodel.FieldDefinition, error) {
	ctx, span := tab.StartSpan(ctx, "newStructField")
	defer span.End()

	// TODO: add the actual struct type name rather than foo
	ident := "fooStruct"

	result := *astmodel.NewFieldDefinition(schema.Property, ident)
	if schema.Description != nil {
		result = result.WithDescription(*schema.Description)
	}

	return &result, nil
}

func newArrayField(ctx context.Context, schema *gojsonschema.SubSchema, arrayType *ast.ArrayType) (*astmodel.FieldDefinition, error) {
	ctx, span := tab.StartSpan(ctx, "newArrayField")
	defer span.End()

	// TODO: add the actual array type name rather than foo
	ident := fmt.Sprintf("[]%s", arrayType.Elt)

	result := *astmodel.NewFieldDefinition(schema.Property, ident)
	if schema.Description != nil {
		result = result.WithDescription(*schema.Description)
	}

	return &result, nil
}

func (scanner *SchemaScanner) objectHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "objectHandler")
	defer span.End()

	objectName := topic.objectName // Default placeholder
	if isObjectName(schema.Property) {
		objectName = schema.Property
	}

	objectTopic := NewObjectScannerTopic(objectName, topic.objectVersion)

	fields, err := scanner.getFields(ctx, objectTopic, schema)
	if err != nil {
		return nil, err
	}

	structDefinition := astmodel.NewStructDefinition(objectName, topic.objectVersion, fields...)

	scanner.Structs = append(scanner.Structs, structDefinition)

	return []astmodel.Definition{
		structDefinition,
	}, nil
}

func (scanner *SchemaScanner) getFields(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]*astmodel.FieldDefinition, error) {
	ctx, span := tab.StartSpan(ctx, "getFields")
	defer span.End()

	fmt.Printf("AST324 STA getFields\n")

	var fields []*astmodel.FieldDefinition
	for _, prop := range schema.PropertiesChildren {
		schemaType, err := getSubSchemaType(prop)
		if _, ok := err.(*UnknownSchemaError); ok {
			// if we don't know the type, we still need to provide the property, we will just provide open interface
			field, err := newField(ctx, prop.Property, "interface{}", schema.Description)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
			continue
		}

		if err != nil {
			return nil, err
		}

		handler := scanner.TypeHandlers[schemaType]
		propDecls, err := handler(ctx, topic, prop)
		if _, ok := err.(*UnknownSchemaError); ok {
			// if we don't know the type, we still need to provide the property, we will just provide open interface
			field, err := newField(ctx, prop.Property, "interface{}", schema.Description)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
			continue
		}

		if err != nil {
			return nil, err
		}

		if len(propDecls) == 0 {
			fmt.Printf("AST0367 WRN No properties found for %s.%s.%s", topic.objectVersion, topic.objectName, topic.propertyName)
			continue
		}

		if isPrimitiveType(schemaType) {
			if len(propDecls) > 1 {
				fmt.Printf("AST0377 WRN Unexpectedly found multiple primitive fields for %s.%s.%s.\n", topic.objectName, topic.objectName, topic.propertyName)
			}

			// Expect to always have a single primitive field
			f := propDecls[0].(*astmodel.FieldDefinition)
			fields = append(fields, f)
			continue
		}

		// allOf or oneOf is left and we expect to have only 1 structure for the field
		if (schemaType == AllOf || schemaType == OneOf || schemaType == AnyOf) && len(propDecls) > 1 {
			// we are not sure what it could be since it's many schemas... interface{}
			field, err := newField(ctx, prop.Property, "interface{}", schema.Description)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
			continue
		}

		// Should only have one declaration left
		decl := propDecls[0]
		switch nt := decl.(type) {
		case *astmodel.FieldDefinition:
			fields = append(fields, decl.(*astmodel.FieldDefinition))
		case *astmodel.StructDefinition:
			// TODO Do we need a custom Fielddefinition that includes a struct?
			fmt.Printf("AST0394 WRN astmodel.StructDefinition not handled\n")
		default:
			fmt.Printf("AST0431 WRN unexpected field type: %T\n", nt)
			// do nothing
		}
	}

	return fields, nil
}

func (scanner *SchemaScanner) refHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "refHandler")
	defer span.End()

	url := schema.Ref.GetUrl()
	if url.Fragment == expressionFragment {
		return []astmodel.Definition{}, nil
	}

	fmt.Printf("AST0455 INF $ref to %s\n", url)

	schemaType, err := getSubSchemaType(schema.RefSchema)
	if err != nil {
		return nil, err
	}

	// If $ref points to an object type, we want to start processing that object definition
	// otherwise we keep our existing topic
	subTopic := topic
	if schemaType == Object {
		n := objectTypeOf(url)
		v := versionOf(url)
		subTopic = NewObjectScannerTopic(n, v)
	}

	handler := scanner.TypeHandlers[schemaType]
	result, err := handler(ctx, subTopic, schema.RefSchema)
	return result, err
}

func (scanner *SchemaScanner) allOfHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "allOfHandler")
	defer span.End()

	var definitions []astmodel.Definition
	for _, all := range schema.AllOf {
		schemaType, err := getSubSchemaType(all)
		if err != nil {
			return nil, err
		}

		handler := scanner.TypeHandlers[schemaType]
		ds, err := handler(ctx, topic, all)
		if err != nil {
			return nil, err
		}

		definitions = append(definitions, ds...)
	}

	return definitions, nil
}

func (scanner *SchemaScanner) oneOfHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "oneOfHandler")
	defer span.End()

	// TODO: Need to handle selecting one of these

	var definitions []astmodel.Definition
	for _, one := range schema.OneOf {
		schemaType, err := getSubSchemaType(one)
		if err != nil {
			return nil, err
		}

		handler := scanner.TypeHandlers[schemaType]
		ds, err := handler(ctx, topic, one)
		if err != nil {
			return nil, err
		}

		definitions = append(definitions, ds...)
	}

	return definitions, nil
}

func (scanner *SchemaScanner) anyOfHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "anyOfHandler")
	defer span.End()

	var definitions []astmodel.Definition
	for _, any := range schema.AnyOf {
		schemaType, err := getSubSchemaType(any)
		if err != nil {
			return nil, err
		}

		handler := scanner.TypeHandlers[schemaType]
		ns, err := handler(ctx, topic, any)
		if err != nil {
			return nil, err
		}

		// if we find a primitive schema, then return it
		if isPrimitiveType(schemaType) {
			return ns, nil
		}

		definitions = append(definitions, ns...)
	}

	// return all possibilities... probably means an interface{}
	return definitions, nil
}

func (scanner *SchemaScanner) arrayHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]astmodel.Definition, error) {
	ctx, span := tab.StartSpan(ctx, "arrayHandler")
	defer span.End()

	if len(schema.ItemsChildren) > 1 {
		return nil, fmt.Errorf("item contains more children than expected: %v", schema.ItemsChildren)
	}

	if len(schema.ItemsChildren) == 0 {
		// there is no type to the elements, so we must assume interface{}
		fmt.Printf("AST0538 WRN Interface assumption unproven")

		field := astmodel.NewFieldDefinition(topic.propertyName, "interface{}")
		return []astmodel.Definition{
			field,
		}, nil
	}

	firstChild := schema.ItemsChildren[0]
	if firstChild.Types.IsTyped() && firstChild.Types.Contains(string(Object)) {
		// this contains an object type, so we need to generate the type
		schemaType, err := getSubSchemaType(firstChild)
		if err != nil {
			return nil, err
		}

		handler := scanner.TypeHandlers[schemaType]
		ds, err := handler(ctx, topic, firstChild)
		if err != nil {
			return nil, err
		}

		// TODO Do we need to create a field here?

		return ds, nil
	}

	// Since the Array Item type is not an object, it should either be a primitive or a anyOf, oneOf or allOf. In that
	// case, we will call to the type handler to build the field and use the field type as the array type.
	schemaType, err := getSubSchemaType(firstChild)
	if err != nil {
		return nil, err
	}

	handler := scanner.TypeHandlers[schemaType]
	definitions, err := handler(ctx, topic, firstChild)
	if err != nil {
		return nil, err
	}

	//TODO: Should not throw away the other elements in nodeList

	if len(definitions) > 1 {
		fmt.Printf("AST0581 WARN discarding extra members of []definitions")
	}

	defn := definitions[0]
	switch defn.(type) {
	case *astmodel.FieldDefinition:
		f := defn.(*astmodel.FieldDefinition)
		field := astmodel.NewFieldDefinition(topic.propertyName, "[]"+f.FieldType())
		return []astmodel.Definition{
			field,
		}, nil

	case *astmodel.StructDefinition:
		// TODO may need a smarter field definition that embeds a custom property
		f := defn.(*astmodel.StructDefinition)
		field := astmodel.NewFieldDefinition(topic.propertyName, f.Name())
		return []astmodel.Definition{
			field,
		}, nil

	default:
		fmt.Printf("AST0640 WRN unexpected definition type (found %T)", defn)
		return []astmodel.Definition{}, nil
	}
}

func getSubSchemaType(schema *gojsonschema.SubSchema) (SchemaType, error) {
	switch {
	case schema.Types.IsTyped() && schema.Types.Contains(string(Object)):
		return Object, nil
	case schema.Types.IsTyped() && schema.Types.Contains(string(String)):
		return String, nil
	case schema.Types.IsTyped() && schema.Types.Contains(string(Number)):
		return Number, nil
	case schema.Types.IsTyped() && schema.Types.Contains(string(Int)):
		return Int, nil
	case schema.Types.IsTyped() && schema.Types.Contains(string(Bool)):
		return Bool, nil
	case schema.Enum != nil:
		return Enum, nil
	case schema.Types.IsTyped() && schema.Types.Contains(string(Array)):
		return Array, nil
	case !schema.Types.IsTyped() && schema.PropertiesChildren != nil:
		return None, nil
	case schema.OneOf != nil:
		return OneOf, nil
	case schema.AllOf != nil:
		return AllOf, nil
	case schema.AnyOf != nil:
		return AnyOf, nil
	case schema.RefSchema != nil:
		return Ref, nil
	default:
		return Unknown, &UnknownSchemaError{
			Schema: schema,
		}
	}
}

func getIdentForPrimitiveType(name SchemaType) (*ast.Ident, error) {
	switch name {
	case String:
		return ast.NewIdent("string"), nil
	case Int:
		return ast.NewIdent("int"), nil
	case Number:
		return ast.NewIdent("float"), nil
	case Bool:
		return ast.NewIdent("bool"), nil
	default:
		return nil, fmt.Errorf("%s is not a simple type and no ast.NewIdent can be created", name)
	}
}

func getPrimitiveType(name SchemaType) (string, error) {
	switch name {
	case String:
		return "string", nil
	case Int:
		return "int", nil
	case Number:
		return "float", nil
	case Bool:
		return "bool", nil
	default:
		return "", fmt.Errorf("%s is not a simple type and no ast.NewIdent can be created", name)
	}
}

func isPrimitiveType(name SchemaType) bool {
	switch name {
	case String, Int, Number, Bool:
		return true
	default:
		return false
	}
}

func asComment(text *string) string {
	if text == nil {
		return ""
	}

	return "// " + *text
}

func isObjectName(name string) bool {
	return name != "$ref" && name != "oneOf"
}

//TODO should this return err or panic?
func objectTypeOf(url *url.URL) string {
	isPathSeparator := func(c rune) bool {
		return c == '/'
	}

	fragmentParts := strings.FieldsFunc(url.Fragment, isPathSeparator)

	return fragmentParts[len(fragmentParts)-1]
}

//TODO should this return err or panic?
func versionOf(url *url.URL) string {
	isPathSeparator := func(c rune) bool {
		return c == '/'
	}

	pathParts := strings.FieldsFunc(url.Path, isPathSeparator)

	return pathParts[1]
}
