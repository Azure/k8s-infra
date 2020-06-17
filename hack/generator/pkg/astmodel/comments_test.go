package astmodel

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestDocumentationCommentFormatting(t *testing.T) {
	g := NewGomegaWithT(t)

	cases := []struct {
		comment       string
		results  []string
	}{
		// Expect short single line comments to be unchanged
		{"foo", []string{"foo"}},
		{"foo", []string{"foo"}},
		{"foo", []string{"foo"}},
	}

	for _, c := range cases {
		lines := formatDocComment(c.comment)
		g.Expect(lines).To(Equal(c.results))
	}
}
