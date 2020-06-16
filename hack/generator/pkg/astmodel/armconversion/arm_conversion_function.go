/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package armconversion

import (
	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"go/ast"
)

// ArmConversionFunction represents an ARM conversion function for converting between an Kubernetes resource
// and an ARM resource.
type ArmConversionFunction struct {
	armTypeName astmodel.TypeName
	armType     *astmodel.ObjectType
	idFactory   astmodel.IdentifierFactory
	toArm       bool
	isResource  bool
}

var _ astmodel.Function = &ArmConversionFunction{}

// RequiredImports returns the imports required for this conversion function
func (c *ArmConversionFunction) RequiredImports() []astmodel.PackageReference {
	var result []astmodel.PackageReference

	result = append(result, c.armType.RequiredImports()...)
	//result = append(result, c.kubeType.RequiredImports()...)
	result = append(result, astmodel.MakeGenRuntimePackageReference())
	result = append(result, astmodel.MakePackageReference("fmt"))

	return result
}

// References this type has to the given type
func (c *ArmConversionFunction) References() astmodel.TypeNameSet {
	return c.armType.References()
}

// AsFunc returns the function as a go ast
func (c *ArmConversionFunction) AsFunc(codeGenerationContext *astmodel.CodeGenerationContext, receiver astmodel.TypeName, methodName string) *ast.FuncDecl {
	if c.toArm {
		return c.asConvertToArmFunc(codeGenerationContext, receiver, methodName)
	} else {
		return c.asConvertFromArmFunc(codeGenerationContext, receiver, methodName)
	}
}

// Equals determines if this function is equal to the passed in function
func (c *ArmConversionFunction) Equals(other astmodel.Function) bool {
	if o, ok := other.(*ArmConversionFunction); ok {
		return c.armType.Equals(o.armType) &&
			c.armTypeName.Equals(o.armTypeName) &&
			c.toArm == o.toArm
	}

	return false
}
