/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

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
		{"Drive", "D:\\", false},
		{"Top level", "/foo/", false},
		{"Top level, Windows", "D:\\foo\\", false},
		{"Nested", "/foo/bar/", false},
		{"Nested, Windows", "D:\\foo\\bar\\", false},
		// Paths to skip
		{"Skip top level", "/examples/", true},
		{"Skip top level, Windows", "D:\\examples\\", true},
		{"Skip nested", "/foo/examples/", true},
		{"Skip nested, Windows", "D:\\foo\\examples\\", true},
		{"Skip nested, trailing directory", "/foo/examples/bar/", true},
		{"Skip nested, trailing directory, Windows", "D:\\foo\\examples\\bar\\", true},
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
