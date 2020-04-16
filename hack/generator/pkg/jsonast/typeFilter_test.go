package jsonast_test

import (
	"testing"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/Azure/k8s-infra/hack/generator/pkg/jsonast"

	. "github.com/onsi/gomega"
)

func Test_FilterByName_CorrectlySelectsStructs(t *testing.T) {
	g := NewGomegaWithT(t)

	person := astmodel.NewStructDefinition("person", "2020-01-01")
	post := astmodel.NewStructDefinition("post", "2019-01-01")
	student := astmodel.NewStructDefinition("student", "2019-01-01")
	filter := jsonast.TypeFilter{Name: "p*"}

	// Name starts with "p" should be selected
	g.Expect(filter.AppliesToType(person)).To(BeTrue())
	g.Expect(filter.AppliesToType(post)).To(BeTrue())

	// Name does not start with "p" should not be selected
	g.Expect(filter.AppliesToType(student)).To(BeFalse())
}

func Test_FilterByVersion_CorrectlySelectsStructs(t *testing.T) {
	g := NewGomegaWithT(t)

	person := astmodel.NewStructDefinition("person", "2020-01-01")
	post := astmodel.NewStructDefinition("post", "2019-01-01")
	student := astmodel.NewStructDefinition("student", "2019-01-01")
	filter := jsonast.TypeFilter{Version: "2019-*"}

	// Version from 2019 should be selected
	g.Expect(filter.AppliesToType(post)).To(BeTrue())
	g.Expect(filter.AppliesToType(student)).To(BeTrue())

	// Version not from 2019 should not be selected
	g.Expect(filter.AppliesToType(person)).To(BeFalse())
}

func Test_FilterByMultipleConditions_CorrectlySelectsStructs(t *testing.T) {
	g := NewGomegaWithT(t)

	person := astmodel.NewStructDefinition("person", "2020-01-01")
	post := astmodel.NewStructDefinition("post", "2019-01-01")
	student := astmodel.NewStructDefinition("student", "2019-01-01")
	filter := jsonast.TypeFilter{Name: "p*", Version: "2019-*"}

	// Version not selected by filter
	g.Expect(filter.AppliesToType(person)).To(BeFalse())

	// Both name and version selected by filter
	g.Expect(filter.AppliesToType(post)).To(BeTrue())

	// Name not selected by filter
	g.Expect(filter.AppliesToType(student)).To(BeFalse())
}
