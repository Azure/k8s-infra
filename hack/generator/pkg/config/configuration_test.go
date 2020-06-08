/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package config

import (
	"testing"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"

	. "github.com/onsi/gomega"
)

// Shared test values:
var package2019 = *astmodel.NewLocalPackageReference("group", "2019-01-01")
var person2019 = astmodel.NewTypeName(package2019, "person")
var post2019 = astmodel.NewTypeName(package2019, "post")
var student2019 = astmodel.NewTypeName(package2019, "student")

var package2020 = *astmodel.NewLocalPackageReference("group", "2020-01-01")
var address2020 = astmodel.NewTypeName(package2020, "address")
var person2020 = astmodel.NewTypeName(package2020, "person")
var professor2020 = astmodel.NewTypeName(package2020, "professor")
var student2020 = astmodel.NewTypeName(package2020, "student")
var tutor2020 = astmodel.NewTypeName(package2020, "tutor")

func Test_WithSingleFilter_FiltersExpectedTypes(t *testing.T) {
	g := NewGomegaWithT(t)
	person := person2020
	post := post2019
	student := student2019

	filter := ExportFilter{Action: ExportFilterActionInclude, Filter: Filter{Version: "2019*"}}
	config := NewConfiguration(&filter)

	g.Expect(config.ShouldExport(person)).To(Equal(Export))
	g.Expect(config.ShouldExport(post)).To(Equal(Export))
	g.Expect(config.ShouldExport(student)).To(Equal(Export))
}

func Test_WithMultipleFilters_FiltersExpectedTypes(t *testing.T) {
	g := NewGomegaWithT(t)
	person := person2020
	post := post2019
	student := student2019
	address := address2020

	versionFilter := ExportFilter{
		Action: ExportFilterActionInclude,
		Filter: Filter{Version: "2019*"},
	}
	nameFilter := ExportFilter{
		Action: ExportFilterActionInclude,
		Filter: Filter{Name: "*ss"},
	}
	config := NewConfiguration(&versionFilter, &nameFilter)

	g.Expect(config.ShouldExport(person)).To(Equal(Export))
	g.Expect(config.ShouldExport(post)).To(Equal(Export))
	g.Expect(config.ShouldExport(student)).To(Equal(Export))
	g.Expect(config.ShouldExport(address)).To(Equal(Export))
}

func Test_WithMultipleFilters_GivesPrecedenceToEarlierFilters(t *testing.T) {
	g := NewGomegaWithT(t)

	alwaysExportPerson := ExportFilter{
		Action: ExportFilterActionInclude,
		Filter: Filter{Name: "person"}}
	exclude2019 := ExportFilter{
		Action: ExportFilterActionExclude,
		Filter: Filter{Version: "2019-01-01"}}
	config := NewConfiguration(&alwaysExportPerson, &exclude2019)

	g.Expect(config.ShouldExport(person2019)).To(Equal(Export))
	g.Expect(config.ShouldExport(student2019)).To(Equal(Skip))

	g.Expect(config.ShouldExport(person2020)).To(Equal(Export))
	g.Expect(config.ShouldExport(professor2020)).To(Equal(Export))
	g.Expect(config.ShouldExport(tutor2020)).To(Equal(Export))
	g.Expect(config.ShouldExport(student2020)).To(Equal(Export))
}
