package jsonast

import (
	"context"
	"errors"
	"fmt"
	"go/ast"

	"github.com/devigned/tab"
	"github.com/xeipuuv/gojsonschema"
)

type (
	SchemaType string

	TypeHandler func(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error)

	BuilderConfig struct {
		TypeHandlers map[SchemaType]TypeHandler
		Filters      []string
	}

	BuilderOption func(cfg *BuilderConfig) error

	UnknownSchemaError struct {
		Schema *gojsonschema.SubSchema
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

// WithTypeHandler will override a default type handler for a given SchemaType. This allows for a consumer to customize
// AST generation.
func WithTypeHandler(schemaType SchemaType, handler TypeHandler) BuilderOption {
	return func(cfg *BuilderConfig) error {
		cfg.TypeHandlers[schemaType] = handler
		return nil
	}
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
func ToNodes(ctx context.Context, schema *gojsonschema.SubSchema, opts ...BuilderOption) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "ToNodes")
	defer span.End()

	cfg := &BuilderConfig{
		TypeHandlers: DefaultTypeHandlers(),
		Filters:      []string{},
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

	rootHandler := cfg.TypeHandlers[schemaType]
	nodes, err := rootHandler(ctx, cfg, schema)
	if err != nil {
		return nil, err
	}

	return nodes, nil
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
		String: stringHandler,
		Int:    intHandler,
		Number: numberHandler,
		Bool:   boolHandler,
		Enum:   enumHandler,
		None:   noneHandler,
	}
}

func enumHandler(ctx context.Context, _ *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "enumHandler")
	defer span.End()

	f, err := newPrimitiveField(ctx, schema, String)
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		f,
	}, nil
}

func noneHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "noneHandler")
	defer span.End()

	fields, err := getFields(ctx, cfg, schema)
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		&ast.FieldList{
			List: fields,
		},
	}, nil
}

func boolHandler(ctx context.Context, _ *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "boolHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, Bool)
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		field,
	}, nil
}

func numberHandler(ctx context.Context, _ *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "numberHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, Number)
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		field,
	}, nil
}

func intHandler(ctx context.Context, _ *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "intHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, Int)
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		field,
	}, nil
}

func stringHandler(ctx context.Context, _ *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "stringHandler")
	defer span.End()

	field, err := newPrimitiveField(ctx, schema, String)
	if err != nil {
		return nil, err
	}

	return []ast.Node{
		field,
	}, nil
}

func newField(ctx context.Context, fieldName string, fieldType ast.Expr, description *string) (*ast.Field, error) {
	ctx, span := tab.StartSpan(ctx, "newField")
	defer span.End()

	var desc string
	if description != nil {
		desc = *description
	}

	doc := &ast.CommentGroup{
		List: []*ast.Comment{
			{
				Text: fmt.Sprintf("%s is %s", fieldName, desc),
			},
		},
	}

	field := &ast.Field{
		Doc: doc,
		Names: []*ast.Ident{
			{
				Name: fieldName,
			},
		},
		Type:    fieldType,
		Tag:     nil, // TODO: add field tags for api hints / json binding
		Comment: nil,
	}

	return field, nil
}

func newPrimitiveField(ctx context.Context, schema *gojsonschema.SubSchema, typeIdent SchemaType) (*ast.Field, error) {
	ctx, span := tab.StartSpan(ctx, "newPrimitiveField")
	defer span.End()

	ident, err := getIdentForPrimitiveType(typeIdent)
	if err != nil {
		return nil, err
	}

	return newField(ctx, schema.Property, ident, schema.Description)

}

func newStructField(ctx context.Context, schema *gojsonschema.SubSchema, structType *ast.StructType) (*ast.Field, error) {
	ctx, span := tab.StartSpan(ctx, "newPrimitiveField")
	defer span.End()

	// TODO: add the actual struct type name rather than foo
	return newField(ctx, schema.Property, ast.NewIdent("foo"), schema.Description)
}

func newArrayField(ctx context.Context, schema *gojsonschema.SubSchema, arrayType *ast.ArrayType) (*ast.Field, error) {
	ctx, span := tab.StartSpan(ctx, "newPrimitiveField")
	defer span.End()

	// TODO: add the actual array type name rather than foo
	return newField(ctx, schema.Property, ast.NewIdent(fmt.Sprintf("[]%s", arrayType.Elt)), schema.Description)
}

func objectHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "objectHandler")
	defer span.End()

	f, err := getFields(ctx, cfg, schema)
	if err != nil {
		return nil, err
	}

	objStruct := &ast.StructType{
		Fields: &ast.FieldList{
			List: f,
		},
	}

	return []ast.Node{
		objStruct,
	}, nil
}

func getFields(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]*ast.Field, error) {
	ctx, span := tab.StartSpan(ctx, "getFields")
	defer span.End()

	var fields []ast.Node
	for _, prop := range schema.PropertiesChildren {
		schemaType, err := getSubSchemaType(prop)
		if _, ok := err.(*UnknownSchemaError); ok {
			// if we don't know the type, we still need to provide the property, we will just provide open interface
			f, err := newField(ctx, prop.Property, ast.NewIdent("interface{}"), schema.Description)
			if err != nil {
				return nil, err
			}
			fields = append(fields, f)
			continue
		}

		if err != nil {
			return nil, err
		}

		propDecls, err := cfg.TypeHandlers[schemaType](ctx, cfg, prop)
		if _, ok := err.(*UnknownSchemaError); ok {
			// if we don't know the type, we still need to provide the property, we will just provide open interface
			f, err := newField(ctx, prop.Property, ast.NewIdent("interface{}"), schema.Description)
			if err != nil {
				return nil, err
			}
			fields = append(fields, f)
			continue
		}

		if err != nil {
			return nil, err
		}

		if isPrimitiveType(schemaType) {
			fields = append(fields, propDecls...)
			continue
		}

		// allOf or oneOf is left and we expect to have only 1 structure for the field
		if (schemaType == AllOf || schemaType == OneOf || schemaType == AnyOf) && len(propDecls) > 1 {
			// we are not sure what it could be since it's many schemas... interface{}
			f, err := newField(ctx, prop.Property, ast.NewIdent("interface{}"), schema.Description)
			if err != nil {
				return nil, err
			}
			fields = append(fields, f)
			continue
		}

		node := propDecls[0]
		switch nt := node.(type) {
		case *ast.Field:
			fields = append(fields, nt)
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

	f := make([]*ast.Field, len(fields))
	for i := 0; i < len(fields); i++ {
		field, ok := fields[i].(*ast.Field)
		if !ok {
			return nil, errors.New("unable to cast ast.Node to field when building struct")
		}
		f[i] = field
	}
	return f, nil
}

func refHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "refHandler")
	defer span.End()

	if schema.Ref.GetUrl().Fragment == expressionFragment {
		return []ast.Node{}, nil
	}

	schemaType, err := getSubSchemaType(schema.RefSchema)
	if err != nil {
		return nil, err
	}

	result, err := cfg.TypeHandlers[schemaType](ctx, cfg, schema.RefSchema)
	return result, err
}

func allOfHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "allOfHandler")
	defer span.End()

	var nodes []ast.Node
	for _, all := range schema.AllOf {
		schemaType, err := getSubSchemaType(all)
		if err != nil {
			return nil, err
		}

		handler := cfg.TypeHandlers[schemaType]
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

func oneOfHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "oneOfHandler")
	defer span.End()

	var decls []ast.Node
	for _, one := range schema.OneOf {
		schemaType, err := getSubSchemaType(one)
		if err != nil {
			return nil, err
		}

		handler := cfg.TypeHandlers[schemaType]
		ds, err := handler(ctx, cfg, one)
		if err != nil {
			return nil, err
		}
		decls = append(decls, ds...)
	}

	return decls, nil
}

func anyOfHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
	ctx, span := tab.StartSpan(ctx, "anyOfHandler")
	defer span.End()

	var nodes []ast.Node
	for _, any := range schema.AnyOf {
		schemaType, err := getSubSchemaType(any)
		if err != nil {
			return nil, err
		}

		ns, err := cfg.TypeHandlers[schemaType](ctx, cfg, any)
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

func arrayHandler(ctx context.Context, cfg *BuilderConfig, schema *gojsonschema.SubSchema) ([]ast.Node, error) {
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

		handler := cfg.TypeHandlers[schemaType]
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

	handler := cfg.TypeHandlers[schemaType]
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

func isPrimitiveType(name SchemaType) bool {
	switch name {
	case String, Int, Number, Bool:
		return true
	default:
		return false
	}
}
