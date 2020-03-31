package jsonast

import (
	"context"
	"fmt"
	"go/ast"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/devigned/tab"
	"github.com/xeipuuv/gojsonschema"
)

type (
	SchemaType string

	TypeHandler func(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error)

	BuilderConfig struct {
		Filters []string
	}

	BuilderOption func(cfg *BuilderConfig) error

	UnknownSchemaError struct {
		Schema *gojsonschema.SubSchema
	}

	// A SchemaScanner is used to scan a JSON Schema extracting and collecting type definitions
	SchemaScanner struct {
		Structs      []*astmodel.StructDefinition
		TypeHandlers map[SchemaType]TypeHandler
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

// WithFilters will apply matching filters on resources in the form of {resource_type}/{api_version}
func WithFilters(filters []string) BuilderOption {
	return func(cfg *BuilderConfig) error {
		cfg.Filters = append(cfg.Filters, filters...)
		return nil
	}
}

// NewSchemaScanner constructs a new scanner, ready for use
func NewSchemaScanner() *SchemaScanner {
	scanner := &SchemaScanner{}
	scanner.TypeHandlers = scanner.DefaultTypeHandlers()
	return scanner
}

// WithTypeHandler will override a default type handler for a given SchemaType. This allows for a consumer to customize
// AST generation.
func (scanner *SchemaScanner) AddTypeHandler(schemaType SchemaType, handler TypeHandler) {
	scanner.TypeHandlers[schemaType] = handler
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
func (scanner *SchemaScanner) ToNodes(ctx context.Context, schema *gojsonschema.SubSchema, opts ...BuilderOption) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "ToNodes")
	defer span.End()

	cfg := &BuilderConfig{
		Filters: []string{},
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	schemaType, err := getSubSchemaType(schema)
	if err != nil {
		return nil, err
	}

	rootHandler := scanner.TypeHandlers[schemaType]
	nodes, err := rootHandler(ctx, cfg, schema)
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

func (scanner *SchemaScanner) enumHandler(ctx context.Context, _ *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "enumHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, String)
	if err != nil {
		return nil, err
	}

	node, err := field.AsAst()
	if err != nil {
		return nil, err
	}
	return []ast.Node{
		node,
	}, nil
}

func (scanner *SchemaScanner) noneHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "noneHandler")
	defer span.End()

	fields, err := scanner.getFields(ctx, cfg, schema)
	if err != nil {
		return nil, err
	}

	fieldList, err := astmodel.ToFieldList(fields)
	if err != nil {
		return nil, err
	}

	return []ast.Node{fieldList}, nil
}

func (scanner *SchemaScanner) boolHandler(ctx context.Context, _ *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "boolHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, Bool)
	if err != nil {
		return nil, err
	}

	node, err := field.AsAst()
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		node,
	}, nil
}

func (scanner *SchemaScanner) numberHandler(ctx context.Context, _ *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "numberHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, Number)
	if err != nil {
		return nil, err
	}

	node, err := field.AsAst()
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		node,
	}, nil
}

func (scanner *SchemaScanner) intHandler(ctx context.Context, _ *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "intHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, Int)
	if err != nil {
		return nil, err
	}

	node, err := field.AsAst()
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		node,
	}, nil
}

func (scanner *SchemaScanner) stringHandler(ctx context.Context, _ *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "stringHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, String)
	if err != nil {
		return nil, err
	}

	node, err := field.AsAst()
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		node,
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

func (scanner *SchemaScanner) objectHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "objectHandler")
	defer span.End()

	fmt.Printf("AST308 STA objectHandler\n")

	fields, err := scanner.getFields(ctx, cfg, schema)
	if err != nil {
		return nil, err
	}

	structDefinition := astmodel.NewStructDefinition(title, fields...)

	scanner.Structs = append(scanner.Structs, structDefinition)

	node, err := structDefinition.AsDeclaration()
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		&node,
	}, nil
}

func (scanner *SchemaScanner) getFields(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]*astmodel.FieldDefinition, error) {
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

		propDecls, err := scanner.TypeHandlers[schemaType](ctx, cfg, prop)
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

		if isPrimitiveType(schemaType) {
			fmt.Printf("AST0377 WRN propDecls not handled\n")
			//fields = append(fields, propDecls...)
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

		// HACK: This check avoids a panic caused by not handling propDecls; remove when fixed
		if propDecls == nil {
			continue
		}

		node := propDecls[0]
		switch nt := node.(type) {
		case *ast.Field:
			fmt.Printf("AST0396 WRN *ast.Field not handled\n")
			//fields = append(fields, nt)
		case *ast.StructType:
			// we have a struct, make a new field
			field, err := newStructField(ctx, schema, nt)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		case *ast.ArrayType:
			field, err := newArrayField(ctx, schema, nt)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		case *ast.FieldList:
			// we have a raw set of fields returned, we need to wrap them into a struct type
			structType := &ast.StructType{
				Fields: nt,
			}
			field, err := newStructField(ctx, schema, structType)
			if err != nil {
				return nil, err
			}
			fields = append(fields, field)
		default:
			return nil, fmt.Errorf("unexpected field type: %+v", nt)
		}
	}

	return fields, nil
}

func (scanner *SchemaScanner) refHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "refHandler")
	defer span.End()

	if schema.Ref.GetUrl().Fragment == expressionFragment {
		return []ast.Node{}, nil
	}

	schemaType, err := getSubSchemaType(schema.RefSchema)
	if err != nil {
		return nil, err
	}

	result, err := scanner.TypeHandlers[schemaType](ctx, cfg, schema.RefSchema)
	return result, err
}

func (scanner *SchemaScanner) allOfHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "allOfHandler")
	defer span.End()

	var nodes []ast.Node
	for _, all := range schema.AllOf {
		schemaType, err := getSubSchemaType(all)
		if err != nil {
			return nil, err
		}

		handler := scanner.TypeHandlers[schemaType]
		ds, err := handler(ctx, cfg, all)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, ds...)
	}

	var fields []*ast.Field
	for _, n := range nodes {
		switch nt := n.(type) {
		case *ast.StructType:
			fields = append(fields, nt.Fields.List...)
		case *ast.FieldList:
			fields = append(fields, nt.List...)
		}
	}

	objStruct := &ast.StructType{
		Fields: &ast.FieldList{
			List: fields,
		},
	}

	return []ast.Node{
		objStruct,
	}, nil
}

func (scanner *SchemaScanner) oneOfHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "oneOfHandler")
	defer span.End()

	var decls []ast.Node
	for _, one := range schema.OneOf {
		schemaType, err := getSubSchemaType(one)
		if err != nil {
			return nil, err
		}

		handler := scanner.TypeHandlers[schemaType]
		ds, err := handler(ctx, cfg, one)
		if err != nil {
			return nil, err
		}
		decls = append(decls, ds...)
	}

	return decls, nil
}

func (scanner *SchemaScanner) anyOfHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "anyOfHandler")
	defer span.End()

	var nodes []ast.Node
	for _, any := range schema.AnyOf {
		schemaType, err := getSubSchemaType(any)
		if err != nil {
			return nil, err
		}

		ns, err := scanner.TypeHandlers[schemaType](ctx, cfg, any)
		if err != nil {
			return nil, err
		}

		// if we find a primitive schema, then return it
		if isPrimitiveType(schemaType) {
			return ns, nil
		}

		nodes = append(nodes, ns...)
	}

	// return all possibilities... probably means an interface{}
	return nodes, nil
}

func (scanner *SchemaScanner) arrayHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "arrayHandler")
	defer span.End()

	if len(schema.ItemsChildren) > 1 {
		return nil, fmt.Errorf("item contains more children than expected: %v", schema.ItemsChildren)
	}

	if len(schema.ItemsChildren) == 0 {
		// there is no type to the elements, so we must assume interface{}
		return []ast.Node{
			&ast.ArrayType{
				Elt: ast.NewIdent("interface{}"),
			},
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
		ds, err := handler(ctx, cfg, firstChild)
		if err != nil {
			return nil, err
		}

		structType := ds[0].(*ast.StructType)

		return []ast.Node{
			&ast.ArrayType{
				Elt: structType,
			},
		}, nil
	}

	// Since the Array Item type is not an object, it should either be a primitive or a anyOf, oneOf or allOf. In that
	// case, we will call to the type handler to build the field and use the field type as the array type.
	schemaType, err := getSubSchemaType(firstChild)
	if err != nil {
		return nil, err
	}

	handler := scanner.TypeHandlers[schemaType]
	nodeList, err := handler(ctx, cfg, firstChild)
	if err != nil {
		return nil, err
	}

	switch nt := nodeList[0].(type) {
	case *ast.StructType:
		return []ast.Node{
			&ast.ArrayType{
				Elt: nt,
			},
		}, nil
	case *ast.Field:
		return []ast.Node{
			&ast.ArrayType{
				Elt: nt.Type,
			},
		}, nil
	case *ast.FieldList:
		return []ast.Node{
			&ast.StructType{
				Fields: nt,
			},
		}, nil
	default:
		return nil, fmt.Errorf("first node was not an *ast.Field, *ast.StructType, *ast.FieldList: %+v", nodeList[0])
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
