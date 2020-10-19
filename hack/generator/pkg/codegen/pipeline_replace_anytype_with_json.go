/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
)

const apiExtensions = "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

// jsonType is the type of fields storing arbitrary JSON content in
// custom resources - apiextensions/v1.JSON.
var jsonType = astmodel.MakeTypeName(
	astmodel.MakeExternalPackageReference(apiExtensions), "JSON",
)

func replaceAnyTypeWithJSON() PipelineStage {
	return MakePipelineStage(
		"replaceAnyTypeWithJSON",
		"Replacing interface{}s with arbitrary JSON",
		func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {
			visitor := astmodel.MakeTypeVisitor()

			visitor.VisitPrimitive = func(_ *astmodel.TypeVisitor, it *astmodel.PrimitiveType, _ interface{}) (astmodel.Type, error) {
				if it == astmodel.AnyType {
					return jsonType, nil
				}
				return it, nil
			}

			results := make(astmodel.Types)
			for _, def := range types {
				d, err := visitor.VisitDefinition(def, nil)
				if err != nil {
					return nil, errors.Wrapf(err, "visiting %q", d.Name())
				}
				results.Add(*d)
			}

			return results, nil
		},
	)
}
