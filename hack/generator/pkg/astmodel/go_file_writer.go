package astmodel

import (
	"bufio"
	"bytes"
	"github.com/Azure/k8s-infra/hack/generator/pkg/astbuilder"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"k8s.io/klog/v2"
	"strings"
)

// GoFileWriter encapsulates the work we do to write a Go AST into a specific file
type GoFileWriter struct {
	name         *ast.Ident
	declarations []ast.Decl
}

func NewGoFileWriter(name *ast.Ident, declarations []ast.Decl) *GoFileWriter {
	return &GoFileWriter{
		name:         name,
		declarations: declarations,
	}
}

// SaveToWriter writes the file to the specifier io.Writer
func (writer GoFileWriter) SaveToWriter(filename string, destination io.Writer) error {

	root := writer.createFileAst()

	// Write generated source into a memory buffer
	fileset := token.NewFileSet()
	fileset.AddFile(filename, 1, 102400)

	var unformattedBuffer bytes.Buffer
	err := format.Node(&unformattedBuffer, fileset, root)
	if err != nil {
		return err
	}

	// This is a nasty technique with only one redeeming characteristic: It works
	reformattedBuffer := writer.addBlankLinesBeforeComments(unformattedBuffer)

	// Read the source from the memory buffer (has the effect similar to 'go fmt')
	var cleanAst ast.Node
	cleanAst, err = parser.ParseFile(fileset, filename, &reformattedBuffer, parser.ParseComments)
	if err != nil {
		klog.Errorf("Failed to reformat code (%s); keeping code as is.", err)
		cleanAst = root
	}

	return format.Node(destination, fileset, cleanAst)
}

func (writer GoFileWriter) createFileAst() *ast.File {

	// We set Package (the offset of the package keyword) so that it follows the header comments
	result := &ast.File{
		Doc:   &ast.CommentGroup{},
		Name:  writer.name,
		Decls: writer.declarations,
	}

	astbuilder.AddComments(&result.Doc.List, CodeGenerationComments)
	astbuilder.AddComment(&result.Doc.List, " Copyright (c) Microsoft Corporation.")
	astbuilder.AddComment(&result.Doc.List, " Licensed under the MIT license.")

	headerLen := astbuilder.CommentLength(result.Doc.List)
	result.Package = token.Pos(headerLen)

	return result
}

// addBlankLinesBeforeComments reads the source in the passed buffer and injects a blank line just
// before each '//' style comment so that the comments are nicely spaced out in the generated code.
func (writer GoFileWriter) addBlankLinesBeforeComments(buffer bytes.Buffer) bytes.Buffer {
	// Read all the lines from the buffer
	var lines []string
	reader := bufio.NewReader(&buffer)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	isComment := func(s string) bool {
		return strings.HasPrefix(strings.TrimSpace(s), "//")
	}

	var result bytes.Buffer
	lastLineWasComment := false
	for _, l := range lines {
		// Add blank line prior to each comment block
		if !lastLineWasComment && isComment(l) {
			result.WriteString("\n")
		}

		result.WriteString(l)
		result.WriteString("\n")

		lastLineWasComment = isComment(l)
	}

	return result
}
