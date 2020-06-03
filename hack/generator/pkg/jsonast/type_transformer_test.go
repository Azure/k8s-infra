/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package jsonast_test

import (
	"testing"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/Azure/k8s-infra/hack/generator/pkg/jsonast"

	. "github.com/onsi/gomega"
)

// Shared test values:
var personTypeName = astmodel.NewTypeName(*astmodel.NewLocalPackageReference("party", "2020-01-01"), "person")
var postTypeName = astmodel.NewTypeName(*astmodel.NewLocalPackageReference("thing", "2019-01-01"), "post")
var studentTypeName = astmodel.NewTypeName(*astmodel.NewLocalPackageReference("role", "2019-01-01"), "student")
var tutorTypeName = astmodel.NewTypeName(*astmodel.NewLocalPackageReference("role", "2019-01-01"), "tutor")

func Test_TransformByGroup_CorrectlySelectsTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	transformer := jsonast.TypeTransformer{
		Group: "role",
		Transform: jsonast.TransformTarget{
			Name: "int",
		},
	}
	err := transformer.Init()
	g.Expect(err).To(BeNil())

	// Roles should be selected
	g.Expect(transformer.TransformTypeName(studentTypeName)).To(Equal(astmodel.IntType))
	g.Expect(transformer.TransformTypeName(tutorTypeName)).To(Equal(astmodel.IntType))

	// Party and Plays should not be selected
	g.Expect(transformer.TransformTypeName(personTypeName)).To(BeNil())
	g.Expect(transformer.TransformTypeName(postTypeName)).To(BeNil())
}

func Test_TransformByVersion_CorrectlySelectsTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	transformer := jsonast.TypeTransformer{
		Version: "2019-*",
		Transform: jsonast.TransformTarget{
			Name: "int",
		},
	}
	err := transformer.Init()
	g.Expect(err).To(BeNil())

	// 2019 versions should be transformed
	g.Expect(transformer.TransformTypeName(studentTypeName)).To(Equal(astmodel.IntType))
	g.Expect(transformer.TransformTypeName(tutorTypeName)).To(Equal(astmodel.IntType))
	g.Expect(transformer.TransformTypeName(postTypeName)).To(Equal(astmodel.IntType))

	// other versions should not
	g.Expect(transformer.TransformTypeName(personTypeName)).To(BeNil())
}

func Test_TransformByName_CorrectlySelectsTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	transformer := jsonast.TypeTransformer{
		Name: "p*",
		Transform: jsonast.TransformTarget{
			Name: "int",
		},
	}
	err := transformer.Init()
	g.Expect(err).To(BeNil())

	// Names starting with p should be transformed
	g.Expect(transformer.TransformTypeName(postTypeName)).To(Equal(astmodel.IntType))
	g.Expect(transformer.TransformTypeName(personTypeName)).To(Equal(astmodel.IntType))

	// other versions should not
	g.Expect(transformer.TransformTypeName(studentTypeName)).To(BeNil())
	g.Expect(transformer.TransformTypeName(tutorTypeName)).To(BeNil())
}

func Test_TransformCanTransform_ToComplexType(t *testing.T) {
	g := NewGomegaWithT(t)

	transformer := jsonast.TypeTransformer{
		Name: "tutor",
		Transform: jsonast.TransformTarget{
			PackagePath: "github.com/Azure/k8s-infra/hack/generator/apis/role/2019-01-01",
			Name:        "student",
		},
	}
	err := transformer.Init()
	g.Expect(err).To(BeNil())

	// Tutor should be student
	g.Expect(transformer.TransformTypeName(tutorTypeName)).To(Equal(studentTypeName))
}
