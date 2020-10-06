package codegen

import (
	"context"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

func injectJsonSerializationTests(idFactory astmodel.IdentifierFactory) PipelineStage {

	return MakePipelineStage(
		"jsonTestCases",
		"Add test cases to verify JSON serialization",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {
			stage := makeInjectJsonSerializationTestsStage(idFactory)
			result := make(astmodel.Types)
			var errs []error
			for n, d := range types {
				updated, err := stage.visitor.VisitDefinition(d, n)
				if err != nil {
					errs = append(errs, err)
				} else {
					result[updated.Name()] = *updated
				}
			}

			if len(errs) > 0 {
				return nil, kerrors.NewAggregate(errs)
			}

			return result, nil
		})
}

type injectJsonSerializationTestsStage struct {
	visitor   astmodel.TypeVisitor
	idFactory astmodel.IdentifierFactory
}

func makeInjectJsonSerializationTestsStage(idFactory astmodel.IdentifierFactory) injectJsonSerializationTestsStage {
	result := injectJsonSerializationTestsStage{
		idFactory: idFactory,
	}

	visitor := astmodel.MakeTypeVisitor()
	//visitor.VisitResourceType = result.visitResourceType
	visitor.VisitObjectType = result.visitObjectType

	result.visitor = visitor
	return result
}

func (s *injectJsonSerializationTestsStage) visitObjectType(
	_ *astmodel.TypeVisitor, objectType *astmodel.ObjectType, ctx interface{}) (astmodel.Type, error) {
	name := ctx.(astmodel.TypeName)
	testcase := astmodel.NewObjectSerializationTestCase(name, objectType, s.idFactory)
	result := objectType.WithTestCase(testcase)
	return result, nil
}

/*
func (s *injectJsonSerializationTestsStage) visitResourceType(
	_ *astmodel.TypeVisitor, resource *astmodel.ResourceType, ctx interface{}) (astmodel.Type, error) {
	name := ctx.(astmodel.TypeName)
	testcase := astmodel.NewObjectSerializationTestCase(resource.SpecType())
	result := resource.WithTestCase(testcase)
	return result, nil
}
*/
