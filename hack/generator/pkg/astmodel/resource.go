/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"
	"go/ast"
	"go/token"
	"k8s.io/klog/v2"

	"github.com/pkg/errors"
)

// ResourceType represents a Kubernetes CRD resource which has both
// spec (the user-requested state) and status (the current state)
type ResourceType struct {
	spec             Type
	status           Type
	isStorageVersion bool
	owner            *TypeName

	// flags are used to identify resources during processing;
	// they have no direct representation in the generated code.
	flags map[Flag]struct{}
}

// NewResourceType defines a new resource type
func NewResourceType(specType Type, statusType Type) *ResourceType {
	return &ResourceType{
		specType, statusType, false, nil, make(map[Flag]struct{})}
}

// NewAzureResourceType defines a new resource type for Azure. It ensures that
// the resource has certain expected properties such as type and name.
// The typeName parameter is just used for logging.
func NewAzureResourceType(specType Type, statusType Type, typeName TypeName) *ResourceType {
	if objectType, ok := specType.(*ObjectType); ok {
		// We have certain expectations about structure for resources
		var nameProperty *PropertyDefinition
		var typeProperty *PropertyDefinition
		isNameOptional := false
		isTypeOptional := false
		for _, property := range objectType.Properties() {
			if property.HasName("Name") {
				nameProperty = property
				if _, ok := property.PropertyType().(*OptionalType); ok {
					isNameOptional = true
				}
			}

			if property.HasName("Type") {
				typeProperty = property
				if _, ok := property.PropertyType().(*OptionalType); ok {
					isTypeOptional = true
				}
			}
		}

		if typeProperty == nil {
			// TODO: These resources are currently missing a type property... We should do something
			// TODO: about that, but for now we just bypass them.
			if typeName.Name() != "EnvironmentsEventSources" &&
				typeName.Name() != "SitesConfig" &&
				typeName.Name() != "SitesSlotsConfig" &&
				typeName.Name() != "ServersAdministrators" {
				panic(fmt.Sprintf("Resource %s is missing type property", typeName))
			}
		}

		if nameProperty == nil {
			klog.V(1).Infof("resource %s is missing field 'Name', fabricating one...", typeName)

			nameProperty = NewPropertyDefinition(PropertyName("Name"), "name", StringType)
			nameProperty.WithDescription("The name of the resource")
			isNameOptional = true
		}

		if isNameOptional {
			// Fix name to be required -- again this is an artifact of bad spec more than anything
			nameProperty = nameProperty.MakeRequired()
			objectType = objectType.WithProperty(nameProperty)
		}

		// If the name is not a string, force it to be -- there are a good number
		// of resources which define name as an enum with a limited set of values.
		// That is actually incorrect because it forbids nested naming from being used
		// (i.e. myresource/mysubresource/enumvalue) and that's the style of naming
		// that we're always using because we deploy each resource standalone.
		if !nameProperty.PropertyType().Equals(StringType) {
			klog.V(4).Infof(
				"Forcing resource %s name property with type %T to be string instead",
				typeName,
				nameProperty.PropertyType())
			nameProperty = nameProperty.WithType(StringType)
			objectType = objectType.WithProperty(nameProperty)
		}

		if isTypeOptional {
			typeProperty = typeProperty.MakeRequired()
			objectType = objectType.WithProperty(typeProperty)
		}
		specType = objectType
	} else {
		klog.Warningf("expected a struct type for resource: %v, got %T", typeName, specType)
		// TODO: handle this better, only Kusto does it
	}

	return NewResourceType(specType, statusType)
}

// assert that ResourceType implements Type
var _ Type = &ResourceType{}

// SpecType returns the type used for specificiation
func (resource *ResourceType) SpecType() Type {
	return resource.spec
}

// StatusType returns the type used for current status
func (resource *ResourceType) StatusType() Type {
	return resource.status
}

// WithStatus returns a new resource that has the specified status type
func (definition *ResourceType) WithStatus(statusType Type) *ResourceType {
	result := *definition
	result.status = statusType
	return &result
}

// AsType converts the ResourceType to go AST Expr
func (resource *ResourceType) AsType(_ *CodeGenerationContext) ast.Expr {
	panic("a resource cannot be used directly as a type")
}

// Equals returns true if the other type is also a ResourceType and has Equal fields
func (resource *ResourceType) Equals(other Type) bool {
	if resource == other {
		return true
	}

	if resource == nil {
		return false
	}

	if otherResource, ok := other.(*ResourceType); ok {

		if otherResource == nil {
			return false
		}

		if !TypeEquals(resource.spec, otherResource.spec) ||
			!TypeEquals(resource.status, otherResource.status) ||
			!resource.isStorageVersion == otherResource.isStorageVersion ||
			len(resource.flags) != len(otherResource.flags) {
			return false
		}

		for f := range resource.flags {
			if _, ok := otherResource.flags[f]; !ok {
				return false
			}
		}

		return true
	}

	return false
}

// References returns the types referenced by Status or Spec parts of the resource
func (resource *ResourceType) References() TypeNameSet {
	spec := resource.spec.References()

	var status TypeNameSet
	if resource.status != nil {
		status = resource.status.References()
	}

	return SetUnion(spec, status)
}

// Owner returns the name of the owner type
func (definition *ResourceType) Owner() *TypeName {
	return definition.owner
}

// MarkAsStorageVersion marks the resource as the Kubebuilder storage version
func (resource *ResourceType) MarkAsStorageVersion() *ResourceType {
	result := resource.copy()
	result.isStorageVersion = true
	return result
}

// WithOwner updates the owner of the resource and returns a copy of the resource
func (definition *ResourceType) WithOwner(owner *TypeName) *ResourceType {
	result := *definition
	result.owner = owner
	return &result
}

// AddFlag includes the specified flag on the resource
func (resource *ResourceType) AddFlag(flag Flag) *ResourceType {
	result := resource.copy()
	result.flags[flag] = struct{}{}
	return result
}

// HasFlag returns true if the specified flag is present on this resource
func (resource *ResourceType) HasFlag(flag Flag) bool {
	_, ok := resource.flags[flag]
	return ok
}

// RequiredImports returns a list of packages required by this
func (resource *ResourceType) RequiredImports() []PackageReference {
	typeImports := resource.spec.RequiredImports()

	if resource.status != nil {
		typeImports = append(typeImports, resource.status.RequiredImports()...)
	}

	typeImports = append(typeImports, MetaV1PackageReference)
	typeImports = append(typeImports, MakeGenRuntimePackageReference())
	typeImports = append(typeImports, MakePackageReference("fmt"))

	return typeImports
}

// AsDeclarations converts the resource type to a set of go declarations
func (resource *ResourceType) AsDeclarations(codeGenerationContext *CodeGenerationContext, name TypeName, description []string) []ast.Decl {

	packageName, err := codeGenerationContext.GetImportedPackageName(MetaV1PackageReference)
	if err != nil {
		panic(errors.Wrapf(err, "resource definition for %s failed to import package", name))
	}

	typeMetaField := defineField("", ast.NewIdent(fmt.Sprintf("%s.TypeMeta", packageName)), "`json:\",inline\"`")
	objectMetaField := defineField("", ast.NewIdent(fmt.Sprintf("%s.ObjectMeta", packageName)), "`json:\"metadata,omitempty\"`")

	/*
		start off with:
			metav1.TypeMeta   `json:",inline"`
			metav1.ObjectMeta `json:"metadata,omitempty"`

		then the Spec/Status properties
	*/
	fields := []*ast.Field{
		typeMetaField,
		objectMetaField,
		defineField("Spec", resource.spec.AsType(codeGenerationContext), "`json:\"spec,omitempty\"`"),
	}

	if resource.status != nil {
		fields = append(fields, defineField("Status", resource.status.AsType(codeGenerationContext), "`json:\"spec,omitempty\"`"))
	}

	resourceIdentifier := ast.NewIdent(name.Name())
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

	if resource.isStorageVersion {
		comments = append(comments, &ast.Comment{
			Text: "// +kubebuilder:storageversion\n",
		})
	}

	addDocComments(&comments, description, 200)

	var declarations []ast.Decl
	resourceDeclaration := &ast.GenDecl{
		Tok:   token.TYPE,
		Specs: []ast.Spec{resourceTypeSpec},
		Doc:   &ast.CommentGroup{List: comments},
	}

	declarations = append(declarations, resourceDeclaration)

	return declarations
}

// copy makes an independent deep clone of this resource
func (resource *ResourceType) copy() *ResourceType {
	result := *resource

	result.flags = make(map[Flag]struct{})
	for k, v := range resource.flags {
		result.flags[k] = v
	}

	return &result
}
