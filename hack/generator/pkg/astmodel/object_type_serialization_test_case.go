/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	ast "github.com/dave/dst"
	"go/token"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
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
	properties := o.makePropertyMap()

	// Find all the simple generators (those with no external dependencies)
	haveSimpleGenerators, simpleGenerators, err := o.createGenerators(properties, genPackageName, genContext, o.createIndependentGenerator)
	if err != nil {
		errs = append(errs, err)
	}

	// Find all the complex generators (dependent on other generators we'll be generating elsewhere)
	haveRelatedGenerators, relatedGenerators, err := o.createGenerators(properties, genPackageName, genContext, o.createRelatedGenerator)
	if err != nil {
		errs = append(errs, err)
	}

	// Remove properties from our runtime
	o.removeByPackage(properties, GenRuntimeReference)

	o.removeByPackage(properties, ApiExtensionsReference)     // TODO: Handle generators for Arbitrary JSON
	o.removeByPackage(properties, ApiExtensionsJsonReference) // TODO: Handle generators for Arbitrary JSON

	// Write errors for any properties we don't handle
	for _, p := range properties {
		errs = append(errs, errors.Errorf("No generator created for %v (%v)", p.PropertyName(), p.PropertyType()))
	}

	var result []ast.Decl

	if !haveSimpleGenerators && !haveRelatedGenerators {
		// No properties that we can generate to test - skip the testing completely
		errs = append(errs, errors.Errorf("No property generators for %v", name))
	} else {
		result = append(result,
			o.createTestRunner(),
			o.createTestMethod(),
			o.createGeneratorDeclaration(genType),
			o.createGeneratorMethod(genPackageName, genType, haveSimpleGenerators, haveRelatedGenerators))

		if haveSimpleGenerators {
			result = append(result, o.createGeneratorsFactoryMethod(o.idOfIndependentGeneratorsFactoryMethod(), simpleGenerators, genType))
		}

		if haveRelatedGenerators {
			result = append(result, o.createGeneratorsFactoryMethod(o.idOfRelatedGeneratorsFactoryMethod(), relatedGenerators, genType))
		}
	}

	if len(errs) > 0 {
		i := "issues"
		if len(errs) == 1 {
			i = "issue"
		}

		klog.Warningf("Encountered %d %s creating JSON Serialisation test for %s", len(errs), i, name)
		for _, err := range errs {
			klog.Warning(err)
		}
	}

	return result
}

func (o ObjectSerializationTestCase) RequiredImports() *PackageImportSet {
	result := NewPackageImportSet()

	// Standard Go Packages
	result.AddImportsOfReferences(JsonReference, ReflectReference, TestingReference)

	// Cmp
	result.AddImportsOfReferences(CmpReference, CmpOptsReference)

	// Gopter
	result.AddImportsOfReferences(GopterReference, GopterGenReference, GopterPropReference)

	// Other References
	result.AddImportOfReference(DiffReference)
	result.AddImportOfReference(PrettyReference)

	for _, prop := range o.objectType.Properties() {
		for _, ref := range prop.PropertyType().RequiredPackageReferences().AsSlice() {
			result.AddImportOfReference(ref)
		}
	}

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

	configureMaxSize := astbuilder.SimpleAssignment(
		&ast.SelectorExpr{
			X:   parameters,
			Sel: ast.NewIdent("MaxSize"),
		},
		token.ASSIGN,
		astbuilder.IntLiteral(10))

	// properties := gopter.NewProperties(parameters)
	defineProperties := astbuilder.SimpleAssignment(
		properties,
		token.DEFINE,
		astbuilder.CallQualifiedFunc("gopter", "NewProperties", ast.NewIdent(parameters)))

	// partial expression: name of the test
	testName := astbuilder.StringLiteralf("Round trip of %v via JSON returns original", o.Subject())

	// partial expression: prop.ForAll(RunTestForX, XGenerator())
	propForAll := astbuilder.CallQualifiedFunc(
		"prop",
		"ForAll",
		ast.NewIdent(o.idOfTestMethod()),
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
		configureMaxSize,
		defineProperties,
		defineTestCase,
		runTests)

	return fn.DefineFunc()
}

func (o ObjectSerializationTestCase) createTestMethod() ast.Decl {
	binId := ast.NewIdent("bin")
	actualId := ast.NewIdent("actual")
	actualFmtId := ast.NewIdent("actualFmt")
	matchId := ast.NewIdent("match")
	subjectId := ast.NewIdent("subject")
	subjectFmtId := ast.NewIdent("subjectFmt")
	resultId := ast.NewIdent("result")
	errId := ast.NewIdent("err")

	// bin, err := json.Marshal(subject)
	serialize := astbuilder.SimpleAssignmentWithErr(
		binId,
		token.DEFINE,
		astbuilder.CallQualifiedFunc("json", "Marshal", ast.NewIdent(subjectId)))

	// if err != nil { return err.Error() }
	serializeFailed := astbuilder.ReturnIfNotNil(errId, astbuilder.CallQualifiedFuncByName("err", "Error"))
		astbuilder.CallQualifiedFunc("err", "Error"))

	// var actual X
	declare := astbuilder.NewVariable(actualId, o.Subject())

	// err = json.Unmarshal(bin, &actual)
	deserialize := astbuilder.SimpleAssignment(
		ast.NewIdent("err"),
		token.ASSIGN,
		astbuilder.CallQualifiedFuncByName("json", "Unmarshal", binId,
			&ast.UnaryExpr{
				Op: token.AND,
				X:  actualId,
			}))

	// if err != nil { return err.Error() }
	deserializeFailed := astbuilder.ReturnIfNotNil(errId, astbuilder.CallQualifiedFuncByName("err", "Error"))
		astbuilder.CallQualifiedFunc("err", "Error"))

	// match := cmp.Equal(subject, actual, cmpopts.EquateEmpty())
	equateEmpty := astbuilder.CallQualifiedFunc("cmpopts", "EquateEmpty")
	compare := astbuilder.SimpleAssignment(
		matchId,
		token.DEFINE,
		astbuilder.CallQualifiedFunc("cmp", "Equal",
			ast.NewIdent(subjectId),
			ast.NewIdent(actualId),
			equateEmpty))

	// if !match { result := diff.Diff(subject, actual); return result }
	prettyPrint := &ast.IfStmt{
		Cond: &ast.UnaryExpr{
			Op: token.NOT,
			X:  matchId,
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				astbuilder.SimpleAssignment(
					actualFmtId,
					token.DEFINE,
					astbuilder.CallQualifiedFunc("pretty", "Sprint", ast.NewIdent(actualId))),
				astbuilder.SimpleAssignment(
					subjectFmtId,
					token.DEFINE,
					astbuilder.CallQualifiedFunc("pretty", "Sprint", ast.NewIdent(subjectId))),
				astbuilder.SimpleAssignment(
					resultId,
					token.DEFINE,
					astbuilder.CallQualifiedFunc("diff", "Diff", ast.NewIdent(subjectFmtId), ast.NewIdent(actualFmtId))),
				astbuilder.Returns(resultId),
			},
		},
	}

	// return ""
	ret := astbuilder.Returns(astbuilder.StringLiteral(""))

	// Create the function
	fn := &astbuilder.FuncDetails{
		Name: o.idOfTestMethod(),
		Returns: []*ast.Field{
			{
				Type: ast.NewIdent("string"),
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

	return fn.DefineFunc()
}

func (o ObjectSerializationTestCase) createGeneratorDeclaration(genType *ast.SelectorExpr) ast.Decl {
	comment := fmt.Sprintf(
		"Generator of %v instances for property testing - lazily instantiated by %v()",
		o.Subject(),
		o.idOfGeneratorMethod(o.subject))

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
		Name: ast.NewIdent(o.idOfGeneratorMethod(o.subject)),
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
			astbuilder.CallQualifiedFunc(
				genPackage,
				"Struct",
				astbuilder.CallQualifiedFunc("reflect", "TypeOf", &ast.CompositeLit{Type: o.Subject()}),
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
			astbuilder.CallQualifiedFunc(
				genPackage,
				"Struct",
				astbuilder.CallQualifiedFunc("reflect", "TypeOf", &ast.CompositeLit{Type: o.Subject()}),
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
	genContext *CodeGenerationContext,
	factory func(name string, propertyType Type, genPackageName string, genContext *CodeGenerationContext) (ast.Expr, error)) (bool, []ast.Stmt, error) {

	gensIdent := ast.NewIdent("gens")

	var handled []PropertyName
	var result []ast.Stmt

	// Iterate over all properties, creating generators where possible
	var errs []error
	for name, prop := range properties {
		g, err := factory(string(name), prop.PropertyType(), genPackageName, genContext)
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

	return len(result) > 0, result, kerrors.NewAggregate(errs)
}

// createIndependentGenerator() will create a generator if the property has a primitive type that
// is directly supported by a Gopter generator, returning nil if the property type isn't supported.
func (o ObjectSerializationTestCase) createIndependentGenerator(
	name string,
	propertyType Type,
	genPackageName string,
	genContext *CodeGenerationContext) (ast.Expr, error) {

	// Handle simple primitive properties
	switch propertyType {
	case StringType:
		return astbuilder.CallQualifiedFunc(genPackageName, "AlphaString"), nil
	case IntType:
		return astbuilder.CallQualifiedFunc(genPackageName, "Int"), nil
	case FloatType:
		return astbuilder.CallQualifiedFunc(genPackageName, "Float32"), nil
	case BoolType:
		return astbuilder.CallQualifiedFunc(genPackageName, "Bool"), nil
	}

	switch t := propertyType.(type) {
	case TypeName:
		types := genContext.GetTypesInCurrentPackage()
		def, ok := types[t]
		if ok {
			return o.createIndependentGenerator(def.Name().name, def.theType, genPackageName, genContext)
		}
		return nil, nil

	case *EnumType:
		return o.createEnumGenerator(name, genPackageName, t)

	case *OptionalType:
		g, err := o.createIndependentGenerator(name, t.Element(), genPackageName, genContext)
		if err != nil {
			return nil, err
		} else if g != nil {
			return astbuilder.CallQualifiedFunc(genPackageName, "PtrOf", g), nil
		}

	case *ArrayType:
		g, err := o.createIndependentGenerator(name, t.Element(), genPackageName, genContext)
		if err != nil {
			return nil, err
		} else if g != nil {
			return astbuilder.CallQualifiedFunc(genPackageName, "SliceOf", g), nil
		}

	case *MapType:
		keyGen, err := o.createIndependentGenerator(name, t.KeyType(), genPackageName, genContext)
		if err != nil {
			return nil, err
		}

		valueGen, err := o.createIndependentGenerator(name, t.ValueType(), genPackageName, genContext)
		if err != nil {
			return nil, err
		}

		if keyGen != nil && valueGen != nil {
			return astbuilder.CallQualifiedFunc(genPackageName, "MapOf", keyGen, valueGen), nil
		}
	}

	// Not a simple property we can handle here
	return nil, nil
}

// createRelatedGenerator() will create a generator if the property has a complex type that is
// defined within the current package, returning nil if the property type isn't supported.
func (o ObjectSerializationTestCase) createRelatedGenerator(
	name string,
	propertyType Type,
	genPackageName string,
	genContext *CodeGenerationContext) (ast.Expr, error) {

	switch t := propertyType.(type) {
	case TypeName:
		_, ok := genContext.GetTypesInPackage(t.PackageReference)
		if ok {
			// This is a type we're defining, so we can create a generator for it
			if t.PackageReference.Equals(genContext.CurrentPackage()) {
				// create a generator for a property referencing a type in this package
				return astbuilder.CallFunc(o.idOfGeneratorMethod(t)), nil
			}

			importName, err := genContext.GetImportedPackageName(t.PackageReference)
			if err != nil {
				return nil, err
			}

			return astbuilder.CallQualifiedFunc(importName, o.idOfGeneratorMethod(t)), nil
		}

		//TODO: Should we invoke a generator for stuff from our runtime package?

		return nil, nil

	case *OptionalType:
		g, err := o.createRelatedGenerator(name, t.Element(), genPackageName, genContext)
		if err != nil {
			return nil, err
		} else if g != nil {
			return astbuilder.CallQualifiedFunc(genPackageName, "PtrOf", g), nil
		}

	case *ArrayType:
		g, err := o.createRelatedGenerator(name, t.Element(), genPackageName, genContext)
		if err != nil {
			return nil, err
		} else if g != nil {
			return astbuilder.CallQualifiedFunc(genPackageName, "SliceOf", g), nil
		}

	case *MapType:
		// We only support primitive types as keys
		keyGen, err := o.createIndependentGenerator(name, t.KeyType(), genPackageName, genContext)
		if err != nil {
			return nil, err
		}

		valueGen, err := o.createRelatedGenerator(name, t.ValueType(), genPackageName, genContext)
		if err != nil {
			return nil, err
		}

		if keyGen != nil && valueGen != nil {
			return astbuilder.CallQualifiedFunc(genPackageName, "MapOf", keyGen, valueGen), nil
		}
	}

	// Not a property we can handle here
	return nil, nil
}

// removeByNamespace() remove all properties with types from the specified namespace
func (o *ObjectSerializationTestCase) removeByPackage(
	properties map[PropertyName]*PropertyDefinition,
	ref PackageReference) {

	// Work out which properties need to be removed because their types come from the specified package
	var toRemove []PropertyName
	for name, prop := range properties {
		propertyType := prop.PropertyType()
		refs := propertyType.RequiredPackageReferences()
		if refs.Contains(ref) {
			toRemove = append(toRemove, name)
		}
	}

	// Remove them
	for _, name := range toRemove {
		delete(properties, name)
	}
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

func (o ObjectSerializationTestCase) idOfGeneratorMethod(typeName TypeName) string {
	name := o.idFactory.CreateIdentifier(
		fmt.Sprintf("%vGenerator", typeName.Name()),
		Exported)
	return name
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

	return astbuilder.CallQualifiedFunc(genPackageName, "OneConstOf", values...), nil
}
