package jsonast

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/devigned/tab"
	"github.com/xeipuuv/gojsonschema"
)

type (
	SchemaType string

	// TypeHandler is a standard delegate used for walking the schema tree
	TypeHandler func(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error)

	UnknownSchemaError struct {
		Schema  *gojsonschema.SubSchema
		Filters []string
	}

	BuilderOption func(scanner *SchemaScanner) error

	// A SchemaScanner is used to scan a JSON Schema extracting and collecting type definitions
	SchemaScanner struct {
		Structs      map[string]*astmodel.StructDefinition
		TypeHandlers map[SchemaType]TypeHandler
		Filters      []string
	}
)

func (scanner *SchemaScanner) FindStruct(name string, version string) (*astmodel.StructDefinition, bool) {
	key := name + "/" + version
	result, ok := scanner.Structs[key]
	return result, ok
}

func (scanner *SchemaScanner) AddStruct(structDefinition *astmodel.StructDefinition) {
	key := structDefinition.Name() + "/" + structDefinition.Version()
	scanner.Structs[key] = structDefinition
}

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
func NewSchemaScanner() *SchemaScanner {
	scanner := &SchemaScanner{
		Structs: make(map[string]*astmodel.StructDefinition),
	}
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
func (scanner *SchemaScanner) ToNodes(ctx context.Context, schema *gojsonschema.SubSchema, opts ...BuilderOption) (astmodel.Definition, error) {
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

	// get initial topic from ID and Title:
	url := schema.ID.GetUrl()
	if schema.Title == nil {
		return nil, fmt.Errorf("Given schema has no Title")
	}

	topicName := *schema.Title
	topicVersion, err := versionOf(url)
	if err != nil {
		return nil, fmt.Errorf("Unable to extract version for schema: %w", err)
	}

	topic := NewObjectScannerTopic(topicName, topicVersion)

	rootHandler := scanner.TypeHandlers[schemaType]
	nodes, err := rootHandler(ctx, topic, schema)
	if err != nil {
		return nil, err
	}

	// TODO: make safer:
	root := astmodel.NewStructDefinition(topic.CreateStructName(), topic.objectVersion, nodes.(*astmodel.StructType).Fields...)
	description := "Generated from: " + url.String()
	root = root.WithDescription(&description)

	scanner.AddStruct(root)

	return root, nil
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
	}
}

func (scanner *SchemaScanner) enumHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "enumHandler")
	defer span.End()

	// if there is an underlying primitive type, return that
	for _, t := range []SchemaType{Bool, Int, Number, String} {
		if schema.Types.Contains(string(t)) {
			return getPrimitiveType(t)
		}
	}

	// assume string
	return astmodel.StringType, nil

	//TODO Create an Enum field that captures the permitted options too
}

func (scanner *SchemaScanner) boolHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "boolHandler")
	defer span.End()

	return astmodel.BoolType, nil
}

func (scanner *SchemaScanner) numberHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "numberHandler")
	defer span.End()

	return astmodel.FloatType, nil
}

func (scanner *SchemaScanner) intHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "intHandler")
	defer span.End()

	return astmodel.IntType, nil
}

func (scanner *SchemaScanner) stringHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "stringHandler")
	defer span.End()

	return astmodel.StringType, nil
}

func (scanner *SchemaScanner) objectHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "objectHandler")
	defer span.End()

	objectName := schema.Property

	objectTopic := NewObjectScannerTopic(objectName, topic.objectVersion)

	fields, err := scanner.getFields(ctx, objectTopic, schema)
	if err != nil {
		return nil, err
	}

	structDefinition := astmodel.NewStructType(fields)
	return structDefinition, nil
}

func (scanner *SchemaScanner) getFields(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) ([]*astmodel.FieldDefinition, error) {
	ctx, span := tab.StartSpan(ctx, "getFields")
	defer span.End()

	log.Printf("STA getFields %v\n", topic)

	var fields []*astmodel.FieldDefinition
	for _, prop := range schema.PropertiesChildren {
		schemaType, err := getSubSchemaType(prop)
		if _, ok := err.(*UnknownSchemaError); ok {
			// if we don't know the type, we still need to provide the property, we will just provide open interface
			field := astmodel.NewFieldDefinition(prop.Property, astmodel.AnyType).WithDescription(schema.Description)
			fields = append(fields, field)
			continue
		}

		if err != nil {
			return nil, err
		}

		propertyTopic := topic.WithProperty(prop.Property)

		handler := scanner.TypeHandlers[schemaType]
		log.Printf("%v: %s", propertyTopic, schemaType)
		propType, err := handler(ctx, propertyTopic, prop)
		if _, ok := err.(*UnknownSchemaError); ok {
			// if we don't know the type, we still need to provide the property, we will just provide open interface
			field := astmodel.NewFieldDefinition(prop.Property, astmodel.AnyType).WithDescription(schema.Description)
			fields = append(fields, field)
			continue
		}

		if err != nil {
			return nil, err
		}

		field := astmodel.NewFieldDefinition(prop.Property, propType).WithDescription(prop.Description)
		fields = append(fields, field)
	}

	// TODO: need to handle additionalProperties
	// store them in a map?

	return fields, nil
}

func (scanner *SchemaScanner) refHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "refHandler")
	defer span.End()

	url := schema.Ref.GetUrl()

	if url.Fragment == expressionFragment {
		return nil, nil
	}

	log.Printf("INF $ref to %s\n", url)

	schemaType, err := getSubSchemaType(schema.RefSchema)
	if err != nil {
		return nil, err
	}

	// make a new topic based on the ref URL
	name, err := objectTypeOf(url)
	if err != nil {
		return nil, err
	}

	version, err := versionOf(url)
	if err != nil {
		return nil, err
	}

	subTopic := NewObjectScannerTopic(name, version)

	if schemaType == Object {
		// see if we already generated a struct for this ref
		// TODO: base this on URL?
		if definition, ok := scanner.FindStruct(subTopic.CreateStructName(), subTopic.objectVersion); ok {
			return &definition.StructReference, nil
		} else {
			// otherwise add a placeholder to avoid recursive calls
			sd := astmodel.NewStructDefinition(subTopic.CreateStructName(), subTopic.objectVersion)
			scanner.AddStruct(sd)
		}
	}

	log.Printf("topic is: %v (%s)\n", subTopic, schemaType)

	handler := scanner.TypeHandlers[schemaType]
	result, err := handler(ctx, subTopic, schema.RefSchema)
	if err != nil {
		return nil, err
	}

	// if we got back a struct type, give it a name
	// (i.e. emit it as a "type X struct {}")
	// and return that instead
	if std, ok := result.(*astmodel.StructType); ok {
		sd := astmodel.NewStructDefinition(
			subTopic.CreateStructName(),
			subTopic.objectVersion,
			std.Fields...,
		)

		description := "Generated from: " + url.String()
		sd = sd.WithDescription(&description)

		// this will overwrite placeholder added above
		scanner.AddStruct(sd)
		return &sd.StructReference, nil
	}

	return result, err
}

func (scanner *SchemaScanner) allOfHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "allOfHandler")
	defer span.End()

	var fields []*astmodel.FieldDefinition
	for _, all := range schema.AllOf {
		schemaType, err := getSubSchemaType(all)
		if err != nil {
			return nil, err
		}

		handler := scanner.TypeHandlers[schemaType]
		d, err := handler(ctx, topic, all)
		if err != nil {
			return nil, err
		}

		// unpack the contents of what we got from subhandlers:
		if d != nil {
			switch d.(type) {
			case *astmodel.StructType:
				// if it's a struct type get all its fields:
				s := d.(*astmodel.StructType)
				fields = append(fields, s.Fields...)

			case *astmodel.StructReference:
				// if it's a reference to a struct type, inherit from it:
				s := d.(*astmodel.StructReference)
				fields = append(fields, astmodel.NewInheritStructDefinition(s))

			default:
				log.Printf("Unhandled type in allOf: %T\n", d)
			}
		}
	}

	result := astmodel.NewStructType(fields)
	return result, nil
}

func (scanner *SchemaScanner) oneOfHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "oneOfHandler")
	defer span.End()

	// make sure we visit everything before bailing out,
	// to get all types generated even if we can't use them
	var results []astmodel.Type
	for _, one := range schema.OneOf {
		schemaType, err := getSubSchemaType(one)
		if err != nil {
			return nil, err
		}

		handler := scanner.TypeHandlers[schemaType]
		result, err := handler(ctx, topic, one)
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

	// bail out, can't handle this yet:
	return astmodel.AnyType, nil
}

func (scanner *SchemaScanner) anyOfHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "anyOfHandler")
	defer span.End()

	// again, make sure we walk everything first
	// to generate types:
	var results []astmodel.Type
	for _, any := range schema.AnyOf {
		schemaType, err := getSubSchemaType(any)
		if err != nil {
			return nil, err
		}

		handler := scanner.TypeHandlers[schemaType]
		result, err := handler(ctx, topic, any)
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

func (scanner *SchemaScanner) arrayHandler(ctx context.Context, topic ScannerTopic, schema *gojsonschema.SubSchema) (astmodel.Type, error) {
	ctx, span := tab.StartSpan(ctx, "arrayHandler")
	defer span.End()

	if len(schema.ItemsChildren) > 1 {
		return nil, fmt.Errorf("item contains more children than expected: %v", schema.ItemsChildren)
	}

	if len(schema.ItemsChildren) == 0 {
		// there is no type to the elements, so we must assume interface{}
		log.Printf("WRN Interface assumption unproven\n")

		return astmodel.NewArrayType(astmodel.AnyType), nil
	}

	// get the only child type and wrap it up as an array type:
	
	onlyChild := schema.ItemsChildren[0]

	schemaType, err := getSubSchemaType(onlyChild)
	if err != nil {
		return nil, err
	}

	handler := scanner.TypeHandlers[schemaType]
	astType, err := handler(ctx, topic, onlyChild)
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

// Extract the name of an object from the supplied schema URL
func objectTypeOf(url *url.URL) (string, error) {
	isPathSeparator := func(c rune) bool {
		return c == '/'
	}

	fragmentParts := strings.FieldsFunc(url.Fragment, isPathSeparator)

	return fragmentParts[len(fragmentParts)-1], nil
}

// Extract the name of an object from the supplied schema URL
func versionOf(url *url.URL) (string, error) {
	isPathSeparator := func(c rune) bool {
		return c == '/'
	}

	pathParts := strings.FieldsFunc(url.Path, isPathSeparator)
	versionRegex, err := regexp.Compile("\\d\\d\\d\\d-\\d\\d-\\d\\d")
	if err != nil {
		return "", fmt.Errorf("Invalid Regex format %w", err)
	}

	for _, p := range pathParts {
		if versionRegex.MatchString(p) {
			return p, nil
		}
	}

	// No version found, that's fine
	return "", nil
}
