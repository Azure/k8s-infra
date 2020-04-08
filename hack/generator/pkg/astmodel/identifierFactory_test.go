package astmodel

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_CreateIdentifier_GivenName_ReturnsExpectedIdentifier(t *testing.T) {
	g := NewGomegaWithT(t)

	CheckStructIdentifier(g, "name", "Name")
	CheckStructIdentifier(g, "Name", "Name")
	CheckStructIdentifier(g, "$schema", "Schema")
	CheckStructIdentifier(g, "my_important_name", "MyImportantName")
	CheckStructIdentifier(g, "MediaServices_liveEvents_liveOutputs_childResource", "MediaServicesLiveEventsLiveOutputsChildResource")
}

func CheckStructIdentifier(g *WithT, name string, expected string) {
	idfactory := NewIdentifierFactory()

	identifier := idfactory.CreateIdentifier(name)
	g.Expect(identifier).To(Equal(expected))
}

func Test_CreateStructIdentifier_WhenIdentifierInvalid_ReturnsError(t *testing.T) {
	g := NewGomegaWithT(t)
	idfactory := NewIdentifierFactory()

	name := idfactory.CreateIdentifier("$bogus!")

	g.Expect(name).To(BeEmpty())
}
