package astmodel

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"github.com/pkg/errors"
	"go/ast"
	"go/token"
	"k8s.io/klog/v2"
)

// ObjectSerializationTestCase represents a test that the object can be losslessly serialized to
// JSON and back again
type ObjectSerializationTestCase struct {
	testName   string
	subject    TypeName
	objectType *ObjectType
	idFactory  IdentifierFactory
}

var _ TestCase = &ObjectSerializationTestCase{}

func NewObjectSerializationTestCase(
	name TypeName,
	objectType *ObjectType,
	idFactory IdentifierFactory) *ObjectSerializationTestCase {
	testName := fmt.Sprintf("%v_WhenSerializedToJson_DeserializesAsEqual", name.Name())
	return &ObjectSerializationTestCase{
		testName:   testName,
		subject:    name,
		objectType: objectType,
		idFactory:  idFactory,
	}
}

func (o ObjectSerializationTestCase) Name() string {
	return o.testName
}

func (o ObjectSerializationTestCase) References() TypeNameSet {
	result := NewTypeNameSet()
	return result
}

func (o ObjectSerializationTestCase) AsFuncs(_ TypeName, genContext *CodeGenerationContext) []ast.Decl {

	gopterPackage, err := genContext.GetImportedPackageName(GopterReference)
	if err != nil {
		panic(errors.Wrap(err, "expected gopter reference to be available"))
	}

	genPackageName, err := genContext.GetImportedPackageName(GopterGenReference)
	if err != nil {
		panic(errors.Wrap(err, "expected gopter/gen reference to be available"))
	}

	genType := &ast.SelectorExpr{
		X:   ast.NewIdent(gopterPackage),
		Sel: ast.NewIdent("Gen"),
	}

	types := genContext.GetTypesInCurrentPackage()

	properties := o.makePropertyMap()
	simpleGenerators := o.createGenerators(properties, genPackageName, types, o.createIndependentGenerator)
	haveSimpleGenerators := len(simpleGenerators) > 0
	relatedGenerators := o.createGenerators(properties, genPackageName, types, o.createRelatedGenerator)
	haveRelatedGenerators := len(relatedGenerators) > 0

	result := []ast.Decl{
		o.createTestRunner(),
		o.createTestMethod(),
		o.createGeneratorDeclaration(genType),
		o.createGeneratorMethod(genPackageName, genType, haveSimpleGenerators, haveRelatedGenerators),
	}

	if haveSimpleGenerators {
		result = append(result, o.createGeneratorsFactoryMethod(o.idOfIndependentGeneratorsFactoryMethod(), simpleGenerators, genType))
	}

	if haveRelatedGenerators {
		result = append(result, o.createGeneratorsFactoryMethod(o.idOfRelatedGeneratorsFactoryMethod(), relatedGenerators, genType))
	}

	return result
}

func (o ObjectSerializationTestCase) RequiredImports() *PackageImportSet {
	result := NewPackageImportSet()
	result.AddImportOfReference(GopterReference)
	result.AddImportOfReference(GopterGenReference)
	result.AddImportOfReference(GopterPropReference)
	result.AddImportOfReference(JsonReference)
	result.AddImportOfReference(PrettyReference)
	result.AddImportOfReference(ReflectReference)
	result.AddImportOfReference(TestingReference)

	return result
}

func (o ObjectSerializationTestCase) Equals(_ TestCase) bool {
	panic("implement me")
}

func (o ObjectSerializationTestCase) createTestRunner() ast.Decl {

	parameters := ast.NewIdent("parameters")
	properties := ast.NewIdent("properties")
	property := ast.NewIdent("Property")
	testingRun := ast.NewIdent("TestingRun")
	t := ast.NewIdent("t")

	// parameters := gopter.DefaultTestParameters()
	defineParameters := astbuilder.SimpleAssignment(
		parameters,
		token.DEFINE,
		astbuilder.CallQualifiedFuncByName("gopter", "DefaultTestParameters"))

	// properties := gopter.NewProperties(parameters)
	defineProperties := astbuilder.SimpleAssignment(
		properties,
		token.DEFINE,
		astbuilder.CallQualifiedFuncByName("gopter", "NewProperties", parameters))

	// partial expression: name of the test
	testName := astbuilder.StringLiteralf("Round trip of %v via JSON returns original", o.Subject())

	// partial expression: prop.ForAll(RunTestForX, XGenerator())
	propForAll := astbuilder.CallQualifiedFuncByName(
		"prop",
		"ForAll",
		o.idOfTestMethod(),
		astbuilder.CallFunc(o.idOfGeneratorMethod(o.subject)))

	// properties.Property("...", prop.ForAll(RunTestForX, XGenerator())
	defineTestCase := astbuilder.InvokeQualifiedFunc(
		properties,
		property,
		testName,
		propForAll)

	// properties.TestingRun(t)
	runTests := astbuilder.InvokeQualifiedFunc(properties, testingRun, t)

	// Define our function
	fn := astbuilder.NewTestFuncDetails(
		o.testName,
		defineParameters,
		defineProperties,
		defineTestCase,
		runTests)

	return fn.DefineFunc()
}

func (o ObjectSerializationTestCase) createTestMethod() ast.Decl {
	binId := ast.NewIdent("bin")
	actualId := ast.NewIdent("actual")
	falseId := ast.NewIdent("false")
	matchId := ast.NewIdent("match")
	subjectId := ast.NewIdent("subject")
	errId := ast.NewIdent("err")

	// bin, err := json.Marshal(subject)
	serialize := astbuilder.SimpleAssignmentWithErr(
		binId,
		token.DEFINE,
		astbuilder.CallMethodByName("json", "Marshal", subjectId))

	// if err == nil { return false }
	serializeFailed := astbuilder.ReturnIfNotNil(errId, falseId)

	// var actual X
	declare := astbuilder.NewVariable(actualId, o.Subject())

	// err = json.Unmarshal(bin, &actual)
	deserialize := astbuilder.SimpleAssignment(
		ast.NewIdent("err"),
		token.ASSIGN,
		astbuilder.CallMethodByName("json", "Unmarshal", binId,
			&ast.UnaryExpr{
				Op: token.AND,
				X:  actualId,
			}))

	// if err != nil { return false }
	deserializeFailed := astbuilder.ReturnIfNotNil(errId, falseId)

	// match := reflect.DeepEqual(subject, actual)
	compare := astbuilder.SimpleAssignment(
		matchId,
		token.DEFINE,
		astbuilder.CallMethodByName("reflect", "DeepEqual", subjectId, actualId))

	// if !match { pretty.Println(subject) }
	prettyPrint := &ast.IfStmt{
		Cond: &ast.UnaryExpr{
			Op: token.NOT,
			X:  matchId,
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				astbuilder.InvokeMethodByName("pretty", "Println", subjectId),
			},
		},
	}

	// return match
	ret := astbuilder.Returns(matchId)

	// Create the function
	fn := &astbuilder.FuncDetails{
		Name: o.idOfTestMethod(),
		Returns: []*ast.Field{
			{
				Type: ast.NewIdent("bool"),
			},
		},
		Body: []ast.Stmt{
			serialize,
			serializeFailed,
			declare,
			deserialize,
			deserializeFailed,
			compare,
			prettyPrint,
			ret,
		},
	}
	fn.AddParameter("subject", o.Subject())
	fn.AddComments(fmt.Sprintf(
		"runs a test to see if a specific instance of %v round trips to JSON and back losslessly",
		o.Subject()))

	return astbuilder.DefineFunc(fn)
}

func (o ObjectSerializationTestCase) createGeneratorDeclaration(genType *ast.SelectorExpr) ast.Decl {
	comment := fmt.Sprintf(
		"Generator of %v instances for property testing - lazily instantiated by %v()",
		o.Subject(),
		o.idOfGeneratorMethod(o.subject).Name)

	decl := astbuilder.VariableDeclaration(
		o.idOfSubjectGeneratorGlobal(),
		genType,
		comment)

	return decl
}

func (o ObjectSerializationTestCase) createGeneratorMethod(
	genPackageName string,
	genType *ast.SelectorExpr,
	haveSimpleGenerators bool,
	haveRelatedGenerators bool) ast.Decl {

	fn := &astbuilder.FuncDetails{
		Name: o.idOfGeneratorMethod(o.subject),
		Returns: []*ast.Field{
			{
				Type: genType,
			},
		},
	}

	fn.AddComments(
		fmt.Sprintf("returns a generator of %v instances for property testing.", o.Subject()),
		fmt.Sprintf("We first initialize %v with a simplified generator based on the fields with primitive types", o.idOfSubjectGeneratorGlobal()),
		"then replacing it with a more complex one that also handles complex fields.",
		"This ensures any cycles in the object graph properly terminate.",
		"The call to gen.Struct() captures the map, so we have to create a new one for the second generator.")

	// If we have already cached our builder, return it immediately
	earlyReturn := astbuilder.ReturnIfNil(
		o.idOfSubjectGeneratorGlobal(),
		o.idOfSubjectGeneratorGlobal())
	fn.AddStatements(earlyReturn)

	if haveSimpleGenerators {
		// Create a simple version of the generator that does not reference generators for related types
		// This serves to terminate any dependency cycles that might occur during creation of a more fully fledged generator

		independentIdent := ast.NewIdent("independentGenerators")
		makeIndependentMap := astbuilder.SimpleAssignment(
			independentIdent,
			token.DEFINE,
			astbuilder.MakeMap(
				ast.NewIdent("string"),
				genType))

		addIndependentGenerators := astbuilder.InvokeFunc(
			o.idOfIndependentGeneratorsFactoryMethod(),
			independentIdent)

		createIndependentGenerator := astbuilder.SimpleAssignment(
			o.idOfSubjectGeneratorGlobal(),
			token.ASSIGN,
			astbuilder.CallQualifiedFuncByName(
				genPackageName,
				"Struct",
				astbuilder.CallQualifiedFuncByName("reflect", "TypeOf", &ast.CompositeLit{Type: o.Subject()}),
				independentIdent))

		fn.AddStatements(makeIndependentMap, addIndependentGenerators, createIndependentGenerator)
	}

	if haveRelatedGenerators {
		// Define a local that contains all the simple generators
		// Have to call the factory method twice as the simple generator above has captured the map;
		// if we reuse or modify the map, chaos ensues.

		allIdent := ast.NewIdent("allGenerators")

		makeAllMap := astbuilder.SimpleAssignment(
			allIdent,
			token.DEFINE,
			astbuilder.MakeMap(
				ast.NewIdent("string"),
				genType))

		addIndependentGenerators := astbuilder.InvokeFunc(
			o.idOfIndependentGeneratorsFactoryMethod(),
			allIdent)

		addRelatedGenerators := astbuilder.InvokeFunc(
			o.idOfRelatedGeneratorsFactoryMethod(),
			allIdent)

		createFullGenerator := astbuilder.SimpleAssignment(
			o.idOfSubjectGeneratorGlobal(),
			token.ASSIGN,
			astbuilder.CallQualifiedFuncByName(
				genPackageName,
				"Struct",
				astbuilder.CallQualifiedFuncByName("reflect", "TypeOf", &ast.CompositeLit{Type: o.Subject()}),
				allIdent))

		fn.AddStatements(makeAllMap, addIndependentGenerators, addRelatedGenerators, createFullGenerator)
	}

	// Return the freshly created (and now cached) generator
	normalReturn := astbuilder.Returns(o.idOfSubjectGeneratorGlobal())
	fn.AddStatements(normalReturn)

	return fn.DefineFunc()
}

func (o ObjectSerializationTestCase) createGeneratorsFactoryMethod(methodName *ast.Ident, generators []ast.Stmt, genType ast.Expr) ast.Decl {

	if len(generators) == 0 {
		// No simple properties, don't generate a method
		return nil
	}

	mapType := &ast.MapType{
		Key:   ast.NewIdent("string"),
		Value: genType,
	}

	fn := &astbuilder.FuncDetails{
		Name: methodName,
		Body: generators,
	}

	fn.AddComments("is a factory method for creating gopter generators")
	fn.AddParameter("gens", mapType)

	return fn.DefineFunc()
}

// createGenerators creates AST fragments for gopter generators to create values for properties
// properties is a map of properties needing generators; properties handled here are removed from the map.
// genPackageName is the name for the gopter/gen package (not hard coded in case it's renamed for conflict resolution)
// factory is a method for creating generators
func (o ObjectSerializationTestCase) createGenerators(
	properties map[PropertyName]*PropertyDefinition,
	genPackageName string,
	types Types,
	factory func(propertyType Type, genPackageName string, types Types) ast.Expr) []ast.Stmt {

	gensIdent := ast.NewIdent("gens")

	var handled []PropertyName
	var result []ast.Stmt

	// Iterate over all properties, creating generators where possible
	for name, prop := range properties {
		g := factory(prop.PropertyType(), genPackageName, types)
		if g != nil {
			insert := astbuilder.InsertMap(
				gensIdent,
				&ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("\"%v\"", prop.PropertyName()),
				},
				g)
			result = append(result, insert)
			handled = append(handled, name)
		}
	}

	// Remove properties we've handled from the map
	for _, name := range handled {
		delete(properties, name)
	}

	return result
}

func (o ObjectSerializationTestCase) createIndependentGenerator(
	propertyType Type,
	genPackageName string,
	types Types) ast.Expr {

	// Handle simple primitive properties
	switch propertyType {
	case StringType:
		return astbuilder.CallQualifiedFuncByName(genPackageName, "AlphaString"), nil
	case IntType:
		return astbuilder.CallQualifiedFuncByName(genPackageName, "Int"), nil
	case FloatType:
		return astbuilder.CallQualifiedFuncByName(genPackageName, "Float32"), nil
	case BoolType:
		return astbuilder.CallQualifiedFuncByName(genPackageName, "Bool"), nil
	}

	switch t := propertyType.(type) {
	case TypeName:
		def, ok := types[t]
		if ok {
			return o.createIndependentGenerator(def.theType, genPackageName, types)
		}

		return nil
	case *EnumType:
		return o.createEnumGenerator(genPackageName, t)
	}

	// Handle optional properties
	if ot, ok := propertyType.(*OptionalType); ok {
		g := o.createIndependentGenerator(ot.Element(), genPackageName, types)
		if g != nil {
			return astbuilder.CallMethodByName(genPackageName, "PtrOf", g)
		}
	}

	// Not a simple property we can handle here
	return nil
}

func (o ObjectSerializationTestCase) createRelatedGenerator(
	propertyType Type,
	genPackageName string,
	types Types) ast.Expr {

	switch t := propertyType.(type) {
	case TypeName:
		_, ok := types[t]
		if ok {
			// Only create a generator for a property referencing a type in this package
			return astbuilder.CallFunc(o.idOfGeneratorMethod(t))
		}

		//TODO: Should we invoke a generator for stuff from our runtime package?

		return nil

	case *OptionalType:
		g := o.createRelatedGenerator(t.Element(), genPackageName, types)
		if g != nil {
			return astbuilder.CallMethodByName(genPackageName, "PtrOf", g)
		}
	}

	// Not a property we can handle here
	return nil
}

// makePropertyMap() makes a map of all the properties on our subject type
func (o *ObjectSerializationTestCase) makePropertyMap() map[PropertyName]*PropertyDefinition {
	result := make(map[PropertyName]*PropertyDefinition)
	for _, prop := range o.objectType.Properties() {
		result[prop.PropertyName()] = prop
	}
	return result
}

func (o ObjectSerializationTestCase) idOfSubjectGeneratorGlobal() *ast.Ident {
	return o.idOfGeneratorGlobal(o.subject)
}

func (o ObjectSerializationTestCase) idOfTestMethod() *ast.Ident {
	id := o.idFactory.CreateIdentifier(
		fmt.Sprintf("RunTestFor%v", o.Subject()),
		Exported)
	return ast.NewIdent(id)
}

func (o ObjectSerializationTestCase) idOfGeneratorGlobal(name TypeName) *ast.Ident {
	id := o.idFactory.CreateIdentifier(
		fmt.Sprintf("cached%vGenerator", name.Name()),
		NotExported)
	return ast.NewIdent(id)
}

func (o ObjectSerializationTestCase) idOfGeneratorMethod(typeName TypeName) *ast.Ident {
	name := o.idFactory.CreateIdentifier(
		fmt.Sprintf("%vGenerator", typeName.Name()),
		Exported)
	return ast.NewIdent(name)
}

func (o ObjectSerializationTestCase) idOfIndependentGeneratorsFactoryMethod() *ast.Ident {
	name := o.idFactory.CreateIdentifier(
		fmt.Sprintf("AddIndependentPropertyGeneratorsFor%v", o.Subject()),
		Exported)
	return ast.NewIdent(name)
}

// idOfRelatedTypesGeneratorsFactoryMethod creates the identifier for the method that creates generators referencing
// other types
func (o ObjectSerializationTestCase) idOfRelatedGeneratorsFactoryMethod() *ast.Ident {
	name := o.idFactory.CreateIdentifier(
		fmt.Sprintf("AddRelatedPropertyGeneratorsFor%v", o.Subject()),
		Exported)
	return ast.NewIdent(name)
}

func (o ObjectSerializationTestCase) Subject() *ast.Ident {
	return ast.NewIdent(o.subject.name)
}

func (o ObjectSerializationTestCase) createEnumGenerator(genPackageName string, enum *EnumType) ast.Expr {
	var kind token.Token

	switch enum.baseType {
	case StringType:
		kind = token.STRING
	case IntType:
		kind = token.INT
	default:
		klog.Warningf("Enum type %v not expected", enum.baseType)
	}

	var values []ast.Expr
	for _, o := range enum.Options() {
		values = append(values, &ast.BasicLit{Value: o.Value, Kind: kind})
	}

	return astbuilder.CallQualifiedFuncByName(genPackageName, "OneConstOf", values...), nil
}
