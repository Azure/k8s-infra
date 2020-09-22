/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog"
)

// simplifyDefinitions creates a pipeline stage that removes any wrapper types prior to actual code generation
func simplifyDefinitions() PipelineStage {
	return MakePipelineStage(
		"simplifyDefinitions",
		"Simplify definitions",
		func(ctx context.Context, defs astmodel.Types) (astmodel.Types, error) {
			visitor := createSimplifyingVisitor()
			var errs []error
			result := make(astmodel.Types)
			for _, def := range defs {
				d, err := visitor.VisitDefinition(def, nil)
				if err != nil {
					errs = append(errs, err)
				} else {
					result[d.Name()] = *d
					if !def.Type().Equals(d.Type()) {
						klog.V(3).Infof("Simplified %v", def.Name())
					}
				}
			}

			if len(errs) > 0 {
				return nil, kerrors.NewAggregate(errs)
			}

			return result, nil
		})
}

func createSimplifyingVisitor() astmodel.TypeVisitor {
	result := astmodel.MakeTypeVisitor()

	result.VisitArmType = func(tv *astmodel.TypeVisitor, at *astmodel.ArmType, ctx interface{}) (astmodel.Type, error) {
		ot := at.ObjectType()
		return tv.Visit(&ot, ctx)
	}

	return result
}
