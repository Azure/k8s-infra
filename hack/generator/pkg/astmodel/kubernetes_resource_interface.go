/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"github.com/pkg/errors"
)

// These are some magical field names which we're going to use or generate
const (
	AzureNameProperty = "AzureName"
	OwnerProperty     = "Owner"
)

// NewArmTransformerImpl creates a new interface with the specified ARM conversion functions
func NewKubernetesResourceInterfaceImpl(
	idFactory IdentifierFactory,
	spec *ObjectType) (*InterfaceImplementation, error) {

	// Check the spec first to ensure it looks how we expect
	ownerProperty := idFactory.CreatePropertyName(OwnerProperty, Exported)
	_, ok := spec.Property(ownerProperty)
	if !ok {
		return nil, errors.Errorf("Resource spec doesn't have %q property", ownerProperty)
	}

	ownerFunc := &objectFunction{
		name:      OwnerProperty,
		o:         spec,
		idFactory: idFactory,
		asFunc:    ownerFunction,
	}

	if !spec.HasFunctionWithName(AzureNameProperty) {
		// no AzureName function, generate one that returns the AzureName property
		azureNameProperty := idFactory.CreatePropertyName(AzureNameProperty, Exported)
		_, ok = spec.Property(azureNameProperty)
		if !ok {
			return nil, errors.Errorf("Resource spec doesn't have %q property", azureNameProperty)
		}

		azureNameFunc := &objectFunction{
			name:      AzureNameProperty,
			o:         spec,
			idFactory: idFactory,
			asFunc:    azureNameFunction,
		}

		return NewInterfaceImplementation(
			MakeTypeName(MakeGenRuntimePackageReference(), "KubernetesResource"),
			ownerFunc,
			azureNameFunc), nil
	} else {
		// already an AzureName function, no need to generate one
		return NewInterfaceImplementation(
			MakeTypeName(MakeGenRuntimePackageReference(), "KubernetesResource"),
			ownerFunc), nil
	}
}

// objectFunction is a simple helper that implements the Function interface. It is intended for use for functions
// that only need information about the object they are operating on
type objectFunction struct {
	name      string
	o         *ObjectType
	idFactory IdentifierFactory

	asFunc func(f *objectFunction, codeGenerationContext *CodeGenerationContext, receiver TypeName, methodName string) *ast.FuncDecl
}

// WithEnumAzureNameFunction adds an AzureName() function that casts an AzureName property
// with an enum value to a string
func (o *ObjectType) WithEnumAzureNameFunction(idFactory IdentifierFactory) *ObjectType {
	azureNameProp, ok := o.properties[AzureNameProperty]
	if !ok {
		panic("WithEnumAzureNameFunction can only be called on a type with an AzureName property")
	}

	return o.WithFunction(&objectFunction{
		name:      "AzureName",
		o:         o,
		idFactory: idFactory,
		asFunc: func(f *objectFunction, codeGenerationContext *CodeGenerationContext, receiver TypeName, methodName string) *ast.FuncDecl {
			receiverIdent := ast.NewIdent(idFactory.CreateIdentifier(receiver.Name(), NotExported))
			receiverType := receiver.AsType(codeGenerationContext)

			fn := &astbuilder.FuncDetails{
				Name:          ast.NewIdent(methodName),
				ReceiverIdent: receiverIdent,
				ReceiverType:  &ast.StarExpr{X: receiverType},
				Body: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.CallExpr{
								// cast from the enum value to string
								Fun: ast.NewIdent("string"),
								Args: []ast.Expr{
									&ast.SelectorExpr{
										X:   receiverIdent,
										Sel: ast.NewIdent(AzureNameProperty),
									},
								},
							},
						},
					},
				},
			}

			fn.AddComments(fmt.Sprintf("returns the Azure name of the resource (string representation of %s)", azureNameProp.PropertyType().String()))
			fn.AddReturns("string")
			return fn.DefineFunc()
		},
	})
}

// WithFixedValueAzureNameFunction adds an AzureName() function that returns a fixed value
func (o *ObjectType) WithFixedValueAzureNameFunction(fixedValue string, idFactory IdentifierFactory) *ObjectType {

	// ensure fixedValue is quoted. This is always the case with enum values we pass,
	// but let's be safe:
	if !(fixedValue[0] == '"' && fixedValue[len(fixedValue)-1] == '"') {
		fixedValue = fmt.Sprintf("%q", fixedValue)
	}

	return o.WithFunction(&objectFunction{
		name:      "AzureName",
		o:         o,
		idFactory: idFactory,
		asFunc: func(f *objectFunction, codeGenerationContext *CodeGenerationContext, receiver TypeName, methodName string) *ast.FuncDecl {
			receiverIdent := ast.NewIdent(idFactory.CreateIdentifier(receiver.Name(), NotExported))
			receiverType := receiver.AsType(codeGenerationContext)

			fn := &astbuilder.FuncDetails{
				Name:          ast.NewIdent(methodName),
				ReceiverIdent: receiverIdent,
				ReceiverType:  &ast.StarExpr{X: receiverType},
				Body: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.BasicLit{
								Kind:  token.STRING,
								Value: fixedValue,
							},
						},
					},
				},
			}

			fn.AddComments(fmt.Sprintf("returns the Azure name of the resource (always %s)", fixedValue))
			fn.AddReturns("string")
			return fn.DefineFunc()
		},
	})
}

var _ Function = &objectFunction{}

// Name returns the unique name of this function
// (You can't have two functions with the same name on the same object or resource)
func (k *objectFunction) Name() string {
	return k.name
}

func (k *objectFunction) RequiredPackageReferences() *PackageReferenceSet {
	// We only require GenRuntime
	return NewPackageReferenceSet(MakeGenRuntimePackageReference())
}

func (k *objectFunction) References() TypeNameSet {
	return k.o.References()
}

func (k *objectFunction) AsFunc(codeGenerationContext *CodeGenerationContext, receiver TypeName) *ast.FuncDecl {
	return k.asFunc(k, codeGenerationContext, receiver, k.name)
}

func (k *objectFunction) Equals(f Function) bool {
	typedF, ok := f.(*objectFunction)
	if !ok {
		return false
	}

	return k.o.Equals(typedF.o)
}

// IsKubernetesResourceProperty returns true if the supplied property name is one of our "magical" names
func IsKubernetesResourceProperty(name PropertyName) bool {
	return name == AzureNameProperty || name == OwnerProperty
}

func ownerFunction(k *objectFunction, codeGenerationContext *CodeGenerationContext, receiver TypeName, methodName string) *ast.FuncDecl {
	receiverIdent := ast.NewIdent(k.idFactory.CreateIdentifier(receiver.Name(), NotExported))
	receiverType := receiver.AsType(codeGenerationContext)

	specSelector := &ast.SelectorExpr{
		X:   receiverIdent,
		Sel: ast.NewIdent("Spec"),
	}

	groupIdent := ast.NewIdent("group")
	kindIdent := ast.NewIdent("kind")

	fn := &astbuilder.FuncDetails{
		Name:          ast.NewIdent(methodName),
		ReceiverIdent: receiverIdent,
		ReceiverType: &ast.StarExpr{
			X: receiverType,
		},
		Params: nil,
		Returns: []*ast.Field{
			{
				Type: &ast.StarExpr{
					X: &ast.SelectorExpr{
						X:   ast.NewIdent(GenRuntimePackageName),
						Sel: ast.NewIdent("ResourceReference"),
					},
				},
			},
		},
		Body: []ast.Stmt{
			lookupGroupAndKindStmt(groupIdent, kindIdent, specSelector),
			&ast.ReturnStmt{
				Results: []ast.Expr{
					createResourceReference(groupIdent, kindIdent, specSelector),
				},
			},
		},
	}

	fn.AddComments("returns the ResourceReference of the owner, or nil if there is no owner")

	return fn.DefineFunc()
}

func lookupGroupAndKindStmt(
	groupIdent *ast.Ident,
	kindIdent *ast.Ident,
	specSelector *ast.SelectorExpr) *ast.AssignStmt {

	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			groupIdent,
			kindIdent,
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent(GenRuntimePackageName),
					Sel: ast.NewIdent("LookupOwnerGroupKind"),
				},
				Args: []ast.Expr{
					specSelector,
				},
			},
		},
	}
}

func createResourceReference(
	groupIdent *ast.Ident,
	kindIdent *ast.Ident,
	specSelector *ast.SelectorExpr) ast.Expr {

	return astbuilder.AddrOf(
		&ast.CompositeLit{
			Type: &ast.SelectorExpr{
				X:   ast.NewIdent(GenRuntimePackageName),
				Sel: ast.NewIdent("ResourceReference"),
			},
			Elts: []ast.Expr{
				&ast.KeyValueExpr{
					Key: ast.NewIdent("Name"),
					Value: &ast.SelectorExpr{
						X: &ast.SelectorExpr{
							X:   specSelector,
							Sel: ast.NewIdent(OwnerProperty),
						},
						Sel: ast.NewIdent("Name"),
					},
				},
				&ast.KeyValueExpr{
					Key:   ast.NewIdent("Group"),
					Value: groupIdent,
				},
				&ast.KeyValueExpr{
					Key:   ast.NewIdent("Kind"),
					Value: kindIdent,
				},
			},
		})
}

func azureNameFunction(k *objectFunction, codeGenerationContext *CodeGenerationContext, receiver TypeName, methodName string) *ast.FuncDecl {
	receiverIdent := ast.NewIdent(k.idFactory.CreateIdentifier(receiver.Name(), NotExported))
	receiverType := receiver.AsType(codeGenerationContext)

	specSelector := &ast.SelectorExpr{
		X:   receiverIdent,
		Sel: ast.NewIdent("Spec"),
	}

	fn := &astbuilder.FuncDetails{
		Name:          ast.NewIdent(methodName),
		ReceiverIdent: receiverIdent,
		ReceiverType: &ast.StarExpr{
			X: receiverType,
		},
		Body: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.SelectorExpr{
						X:   specSelector,
						Sel: ast.NewIdent(AzureNameProperty),
					},
				},
			},
		},
	}

	fn.AddComments("returns the Azure name of the resource")
	fn.AddReturns("string")
	return fn.DefineFunc()
}
