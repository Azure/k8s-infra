/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package config_test

import (
	"github.com/Azure/k8s-infra/hack/generator/pkg/config"
	"testing"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	. "github.com/onsi/gomega"
)

// Shared test values -- note that these are used by type_transformer_test.go too
var person2020 = astmodel.NewTypeName(*astmodel.NewLocalPackageReference("party", "2020-01-01"), "person")
var post2019 = astmodel.NewTypeName(*astmodel.NewLocalPackageReference("thing", "2019-01-01"), "post")
var student2019 = astmodel.NewTypeName(*astmodel.NewLocalPackageReference("role", "2019-01-01"), "student")
var tutor2019 = astmodel.NewTypeName(*astmodel.NewLocalPackageReference("role", "2019-01-01"), "tutor")

func Test_FilterByGroup_CorrectlySelectsStructs(t *testing.T) {
	g := NewGomegaWithT(t)

	filter := config.Filter{Group: "role"}

	// Roles should be selected
	g.Expect(filter.AppliesToType(student2019)).To(BeTrue())
	g.Expect(filter.AppliesToType(tutor2019)).To(BeTrue())

	// Party and Plays should not be selected
	g.Expect(filter.AppliesToType(person2020)).To(BeFalse())
	g.Expect(filter.AppliesToType(post2019)).To(BeFalse())
}

func Test_FilterByVersion_CorrectlySelectsStructs(t *testing.T) {
	g := NewGomegaWithT(t)

	filter := config.Filter{Version: "2019-*"}

	// Version from 2019 should be selected
	g.Expect(filter.AppliesToType(post2019)).To(BeTrue())
	g.Expect(filter.AppliesToType(student2019)).To(BeTrue())

	// Version not from 2019 should not be selected
	g.Expect(filter.AppliesToType(person2020)).To(BeFalse())
}

func Test_FilterByName_CorrectlySelectsStructs(t *testing.T) {
	g := NewGomegaWithT(t)

	filter := config.Filter{Name: "p*"}

	// Name starts with "p" should be selected
	g.Expect(filter.AppliesToType(person2020)).To(BeTrue())
	g.Expect(filter.AppliesToType(post2019)).To(BeTrue())

	// Name does not start with "p" should not be selected
	g.Expect(filter.AppliesToType(student2019)).To(BeFalse())
}

func Test_FilterByMultipleConditions_CorrectlySelectsStructs(t *testing.T) {
	g := NewGomegaWithT(t)

	filter := config.Filter{Name: "p*", Version: "2019-*"}

	// Version not selected by filter
	g.Expect(filter.AppliesToType(person2020)).To(BeFalse())

	// Both name and version selected by filter
	g.Expect(filter.AppliesToType(post2019)).To(BeTrue())

	// Name not selected by filter
	g.Expect(filter.AppliesToType(student2019)).To(BeFalse())
}
