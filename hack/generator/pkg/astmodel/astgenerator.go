package astmodel

import "go/ast"

// AstGenerator represents the ability to generate an Abstract Syntax Tree (AST) node
type AstGenerator interface {
	AsAst() (ast.Node, error)
}
