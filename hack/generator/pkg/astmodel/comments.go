package astmodel

import (
	"go/ast"
	"strings"
)

// Utility methods for adding comments

func addDocComment(commentList *[]*ast.Comment, comment string) {
	if *commentList == nil {
		// if first comment, add a blank comment prior
		*commentList = []*ast.Comment{
			&ast.Comment{
				Text: "\n//",
			},
		}
	}

	for _, c := range formatDocComment(comment) {
		line := c
		if !strings.HasPrefix(line, "//") {
			line = "//"+line
		}

		*commentList = append(*commentList, &ast.Comment{
			Text: line,
		})
	}
}

// formatDocComment splits the supplied comment string up ready for use as a documentation comment
func formatDocComment(comment string) []string {
	var results []string

	results = append(results, comment)
	return results
}
