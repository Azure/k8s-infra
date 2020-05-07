package astmodel

import (
	"go/ast"
	"go/token"
	"sort"
)

// EnumDefinition generates the full definition of an enumeration
type EnumDefinition struct {
	EnumType
}

var _ Definition = (*EnumDefinition)(nil)

func (enum *EnumDefinition) FileNameHint() string {
	return enum.name
}

func (enum *EnumDefinition) Reference() *DefinitionName {
	return &enum.DefinitionName
}

func (enum *EnumDefinition) Type() Type {
	return &enum.EnumType
}

// AsDeclarations generates the Go code representing this definition
func (enum *EnumDefinition) AsDeclarations() []ast.Decl {
	result := []ast.Decl{enum.createBaseDeclaration()}

	for _, v := range enum.Options {
		decl := enum.createValueDeclaration(v)
		result = append(result, decl)
	}

	return result
}

// Tidy does cleanup to ensure deterministic code generation
func (enum *EnumDefinition) Tidy() {
	sort.Slice(enum.Options, func(left int, right int) bool {
		return enum.Options[left].Identifier < enum.Options[right].Identifier
	})
}

func (enum *EnumDefinition) createBaseDeclaration() ast.Decl {
	var identifier *ast.Ident
	identifier = ast.NewIdent(enum.name)

	typeSpecification := &ast.TypeSpec{
		Name: identifier,
		Type: enum.BaseType.AsType(),
	}

	declaration := &ast.GenDecl{
		Tok: token.TYPE,
		Doc: &ast.CommentGroup{},
		Specs: []ast.Spec{
			typeSpecification,
		},
	}

	return declaration
}

func (enum *EnumDefinition) createValueDeclaration(value EnumValue) ast.Decl {

	var enumIdentifier *ast.Ident
	enumIdentifier = ast.NewIdent(enum.name)

	valueIdentifier := ast.NewIdent(enum.Name() + "_" + value.Identifier)
	valueLiteral := ast.BasicLit{
		Kind:  token.STRING,
		Value: value.Value,
	}

	valueSpec := &ast.ValueSpec{
		Names: []*ast.Ident{valueIdentifier},
		Values: []ast.Expr{
			&ast.CallExpr{
				Fun:  enumIdentifier,
				Args: []ast.Expr{&valueLiteral},
			},
		},
	}

	declaration := &ast.GenDecl{
		Tok: token.CONST,
		Doc: &ast.CommentGroup{},
		Specs: []ast.Spec{
			valueSpec,
		},
	}

	return declaration
}
