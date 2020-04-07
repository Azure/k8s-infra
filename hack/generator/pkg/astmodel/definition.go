package astmodel

import "go/ast"

// Definition represents models that can render into Go code
type Definition interface {
	// AsAst() renders a definition into a Go abstract syntax tree
	AsAst() ast.Node
}

