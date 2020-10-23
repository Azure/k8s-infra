/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"github.com/pkg/errors"
	"go/ast"
	"go/token"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog"
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

func (o ObjectSerializationTestCase) AsFuncs(name TypeName, genContext *CodeGenerationContext) []ast.Decl {

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

	var errs []error
	types := genContext.GetTypesInCurrentPackage()

	properties := o.makePropertyMap()
	simpleGenerators, err := o.createGenerators(properties, genPackageName, types, o.createIndependentGenerator)
	if err != nil {
		errs = append(errs, err)
	}

	haveSimpleGenerators := len(simpleGenerators) > 0

	relatedGenerators, err := o.createGenerators(properties, genPackageName, types, o.createRelatedGenerator)
	if err != nil {
		errs = append(errs, err)
	}

	haveRelatedGenerators := len(relatedGenerators) > 0

	// TODO - enable this once we reduce the noise
	// TODO - work out how to create a generator for a type in a different package
	for _, p := range properties {
		errs = append(errs, errors.Errorf("No generator created for %v (%v)", p.PropertyName(), p.PropertyType()))
	}

	if !haveSimpleGenerators && !haveRelatedGenerators {
		// No properties that we can generate to test - skip the testing completely
		errs = append(errs, errors.Errorf("No property generators for %v", name))
	}

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

	if len(errs) > 0 {
		klog.Warningf("Encountered %v issues creating JSON Serialisation test for %v", len(errs), name)
		for _, err := range errs {
			klog.Warning(err)
		}
	}

	return result
}

func (o ObjectSerializationTestCase) RequiredImports() *PackageImportSet {
	result := NewPackageImportSet()
	result.AddImportOfReference(FmtReference)
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

	// if !match { pretty.Println(subject); pretty.Println(actual) }
	prettyPrint := &ast.IfStmt{
		Cond: &ast.UnaryExpr{
			Op: token.NOT,
			X:  matchId,
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				astbuilder.InvokeMethodByName("fmt", "Println", astbuilder.LiteralString("===== Subject =====")),
				astbuilder.InvokeMethodByName("pretty", "Println", subjectId),
				astbuilder.InvokeMethodByName("fmt", "Println", astbuilder.LiteralString("===== Actual =====")),
				astbuilder.InvokeMethodByName("pretty", "Println", actualId),
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
	earlyReturn := astbuilder.ReturnIfNotNil(
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

		fn.AddStatements(makeAllMap)

		if haveSimpleGenerators {
			addIndependentGenerators := astbuilder.InvokeFunc(
				o.idOfIndependentGeneratorsFactoryMethod(),
				allIdent)
			fn.AddStatements(addIndependentGenerators)
		}

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

		fn.AddStatements(addRelatedGenerators, createFullGenerator)
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
	factory func(name string, propertyType Type, genPackageName string, types Types) (ast.Expr, error)) ([]ast.Stmt, error) {

	gensIdent := ast.NewIdent("gens")

	var handled []PropertyName
	var result []ast.Stmt

	// Iterate over all properties, creating generators where possible
	var errs []error
	for name, prop := range properties {
		g, err := factory(string(name), prop.PropertyType(), genPackageName, types)
		if err != nil {
			errs = append(errs, err)
		} else if g != nil {
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

	return result, kerrors.NewAggregate(errs)
}

func (o ObjectSerializationTestCase) createIndependentGenerator(
	name string,
	propertyType Type,
	genPackageName string,
	types Types) (ast.Expr, error) {

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
			return o.createIndependentGenerator(def.Name().name, def.theType, genPackageName, types)
		}
		return nil, nil

	case *EnumType:
		return o.createEnumGenerator(name, genPackageName, t)

	case *OptionalType:
		g, err := o.createIndependentGenerator(name, t.Element(), genPackageName, types)
		if err != nil {
			return nil, err
		} else if g != nil {
			return astbuilder.CallQualifiedFuncByName(genPackageName, "PtrOf", g), nil
		}

	case *ArrayType:
		g, err := o.createIndependentGenerator(name, t.Element(), genPackageName, types)
		if err != nil {
			return nil, err
		} else if g != nil {
			return astbuilder.CallQualifiedFuncByName(genPackageName, "SliceOf", g), nil
		}

	case *MapType:
		keyGen, err := o.createIndependentGenerator(name, t.KeyType(), genPackageName, types)
		if err != nil {
			return nil, err
		}

		valueGen, err := o.createIndependentGenerator(name, t.ValueType(), genPackageName, types)
		if err != nil {
			return nil, err
		}

		if keyGen != nil && valueGen != nil {
			return astbuilder.CallQualifiedFuncByName(genPackageName, "MapOf", keyGen, valueGen), nil
		}
	}

	// Not a simple property we can handle here
	return nil, nil
}

func (o ObjectSerializationTestCase) createRelatedGenerator(
	name string,
	propertyType Type,
	genPackageName string,
	types Types) (ast.Expr, error) {

	switch t := propertyType.(type) {
	case TypeName:
		_, ok := types[t]
		if ok {
			// Only create a generator for a property referencing a type in this package
			return astbuilder.CallFunc(o.idOfGeneratorMethod(t)), nil
		}

		//TODO: Should we invoke a generator for stuff from our runtime package?

		return nil, nil

	case *OptionalType:
		g, err := o.createRelatedGenerator(name, t.Element(), genPackageName, types)
		if err != nil {
			return nil, err
		} else if g != nil {
			return astbuilder.CallQualifiedFuncByName(genPackageName, "PtrOf", g), nil
		}

	case *ArrayType:
		g, err := o.createRelatedGenerator(name, t.Element(), genPackageName, types)
		if err != nil {
			return nil, err
		} else if g != nil {
			return astbuilder.CallQualifiedFuncByName(genPackageName, "SliceOf", g), nil
		}
	}

	// Not a property we can handle here
	return nil, nil
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

func (o ObjectSerializationTestCase) createEnumGenerator(enumName string, genPackageName string, enum *EnumType) (ast.Expr, error) {
	var values []ast.Expr
	for _, o := range enum.Options() {
		id := GetEnumValueId(enumName, o)
		values = append(values, ast.NewIdent(id))
	}

	return astbuilder.CallQualifiedFuncByName(genPackageName, "OneConstOf", values...), nil
}
