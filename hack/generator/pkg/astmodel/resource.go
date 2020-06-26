/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"go/ast"
	"go/token"
)

// CreateResourceDefinitions creates definitions for a resource
func CreateResourceDefinitions(name *TypeName, specType *StructType, statusType *StructType, idFactory IdentifierFactory) (TypeDefiner, []TypeDefiner) {

	var others []TypeDefiner

	var specName *TypeName
	if specType != nil {
		specName = NewTypeName(name.PackageReference, name.Name()+"Spec")
		specDef, specOthers := specType.CreateDefinitions(specName, idFactory)

		others = append(append(others, specDef), specOthers...)
	} else {
		panic("spec must always be provided")
	}

	var statusName *TypeName
	if statusType != nil {
		statusName = NewTypeName(name.PackageReference, name.Name()+"Status")
		statusDef, statusOthers := statusType.CreateDefinitions(statusName, idFactory)

		others = append(append(others, statusDef), statusOthers...)
	}

	this := &ResourceDefinition{typeName: name, spec: specName, status: statusName, isStorageVersion: false}

	return this, others
}

// ResourceDefinition represents an ARM resource
type ResourceDefinition struct {
	typeName         *TypeName
	spec             *TypeName
	status           *TypeName
	isStorageVersion bool
	description      *string
}

// assert that ResourceDefinition implements TypeDefiner
var _ TypeDefiner = &ResourceDefinition{}

func (definition *ResourceDefinition) Name() *TypeName {
	return definition.typeName
}

func (definition *ResourceDefinition) Type() Type {
	return definition.spec // TODO?????
}

func (definition *ResourceDefinition) WithIsStorageVersion(isStorageVersion bool) *ResourceDefinition {
	result := *definition
	result.isStorageVersion = isStorageVersion
	return &result
}

func (definition *ResourceDefinition) WithDescription(description *string) TypeDefiner {
	result := *definition
	result.description = description
	return &result
}

// TODO: metav1 import should be added via RequiredImports?
var typeMetaField = defineField("", "metav1.TypeMeta", "`json:\",inline\"`")
var objectMetaField = defineField("", "metav1.ObjectMeta", "`json:\"metadata,omitempty\"`")

// AsDeclarations converts the ResourceDefinition to a go declaration
func (definition *ResourceDefinition) AsDeclarations(codeGenerationContext *CodeGenerationContext) []ast.Decl {

	/*
		start off with:
			metav1.TypeMeta   `json:",inline"`
			metav1.ObjectMeta `json:"metadata,omitempty"`

		then the Spec/Status fields
	*/
	fields := []*ast.Field{
		typeMetaField,
		objectMetaField,
		defineField("Spec", definition.spec.name, "`json:\"spec,omitempty\"`"),
	}

	if definition.status != nil {
		fields = append(fields, defineField("Status", definition.status.name, "`json:\"spec,omitempty\"`"))
	}

	resourceIdentifier := ast.NewIdent(definition.typeName.name)
	resourceTypeSpec := &ast.TypeSpec{
		Name: resourceIdentifier,
		Type: &ast.StructType{
			Fields: &ast.FieldList{List: fields},
		},
	}

	comments :=
		[]*ast.Comment{
			{
				Text: "// +kubebuilder:object:root=true\n",
			},
		}

	if definition.isStorageVersion {
		comments = append(comments, &ast.Comment{
			Text: "// +kubebuilder:storageversion\n",
		})
	}

	if definition.description != nil {
		comments = append(comments, &ast.Comment{
			Text: "/*" + *definition.description + "*/",
		})
	}

	return []ast.Decl{
		&ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{resourceTypeSpec},
			Doc:   &ast.CommentGroup{List: comments},
		},
	}
}
