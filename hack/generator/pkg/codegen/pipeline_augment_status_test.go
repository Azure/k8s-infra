package codegen

import (
	. "github.com/onsi/gomega"

	"testing"
)

func Test_ShouldSkipDir_GivenPath_HasExpectedResult(t *testing.T) {
	cases := []struct {
		name       string
		path       string
		shouldSkip bool
	}{
		// Simple paths
		{"Root", "/", false},
		{"Top level", "/foo/", false},
		{"Nested", "/foo/bar/", false},
		// Paths to skip
		{"Skip top level", "/examples/", true},
		{"Skip nested", "/foo/examples/", true},
		{"Skip nested, trailing directory", "/foo/examples/bar/", true},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			skipped := shouldSkipDir(c.path)

			g.Expect(skipped).To(Equal(c.shouldSkip))
		})
	}
}
