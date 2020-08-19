/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
)

// NewArmTransformerImpl creates a new interface with the specified ARM conversion functions
func NewKubernetesResourceInterfaceImpl(
	resource *ResourceType) *InterfaceImplementation { // TODO: Is taking resource here right...?

	//funcs := map[string]astmodel.Function{
	//	"ConvertToArm": &ArmConversionFunction{
	//		armTypeName: armTypeName,
	//		armType:     armType,
	//		idFactory:   idFactory,
	//		direction:   ConversionDirectionToArm,
	//		isResource:  isResource,
	//	},
	//	"PopulateFromArm": &ArmConversionFunction{
	//		armTypeName: armTypeName,
	//		armType:     armType,
	//		idFactory:   idFactory,
	//		direction:   ConversionDirectionFromArm,
	//		isResource:  isResource,
	//	},
	//}
	//
	//result := astmodel.NewInterfaceImplementation(
	//	astmodel.MakeTypeName(astmodel.MakeGenRuntimePackageReference(), "ArmTransformer"),
	//	funcs)
	//
	//return result

	return nil
}

type KubernetesResourceOwnerFunction struct {
}

func (k KubernetesResourceOwnerFunction) RequiredImports() []PackageReference {
	panic("implement me")
}

func (k KubernetesResourceOwnerFunction) References() TypeNameSet {
	panic("implement me")
}

func (k KubernetesResourceOwnerFunction) AsFunc(codeGenerationContext *CodeGenerationContext, receiver TypeName, methodName string) *ast.FuncDecl {
	panic("implement me")
}

func (k KubernetesResourceOwnerFunction) Equals(f Function) bool {
	panic("implement me")
}

var _ Function = &KubernetesResourceOwnerFunction{}
