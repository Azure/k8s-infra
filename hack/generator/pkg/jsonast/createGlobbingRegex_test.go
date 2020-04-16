package jsonast

import (
	"testing"

	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

func Test_CreateGlobbingRegex_ReturnsExpectedRegex(t *testing.T) {
	g := NewGomegaWithT(t)
	CheckReturnsExpectedRegex(g, "*preview", "^.*preview$")
	CheckReturnsExpectedRegex(g, "*.bak", "^.*\\.bak$")
	CheckReturnsExpectedRegex(g, "2014*", "^2014.*$")
	CheckReturnsExpectedRegex(g, "2014-??-??", "^2014-..-..$")
}

func CheckReturnsExpectedRegex(g *gomega.WithT, globbing string, regex string) {
	r := createGlobbingRegex(globbing)
	s := r.String()
	g.Expect(s).To(Equal(regex))
}

func Test_GlobbingRegex_MatchesExpectedStrings(t *testing.T) {
	g := NewGomegaWithT(t)
	CheckMatchFound(g, "*preview", "2020-02-01preview")
	CheckMatchNotFound(g, "*preview", "2020-02-01")

	CheckMatchFound(g, "2020-*", "2020-02-01")
	CheckMatchFound(g, "2020-*", "2020-01-01")
	CheckMatchNotFound(g, "2020-*", "2019-01-01")
	CheckMatchNotFound(g, "2020-*", "2015-07-01")
}

func CheckMatchFound(g *gomega.WithT, globbing string, candidate string) {
	r := createGlobbingRegex(globbing)
	match := r.MatchString(candidate)
	g.Expect(match).To(BeTrue())
}

func CheckMatchNotFound(g *gomega.WithT, globbing string, candidate string) {
	r := createGlobbingRegex(globbing)
	match := r.MatchString(candidate)
	g.Expect(match).To(BeFalse())
}
