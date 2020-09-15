/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package controllers

import (
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"testing"

	// TODO: Do we want to use a sample object rather than a code generated one?
	batch "github.com/Azure/k8s-infra/hack/generated/apis/microsoft.batch/v20170901"

	. "github.com/onsi/gomega"
)

func Test_ResourceSpecToArmResourceSpec(t *testing.T) {
	// g := NewGomegaWithT(t)

	// TODO: Write this test
}

func createDummyResource() *batch.BatchAccount {
	return &batch.BatchAccount{
		Spec: batch.BatchAccounts_Spec{
			ApiVersion: "apiVersion",
			AzureName:  "azureName",
			Location:   "westus",
			Owner:      genruntime.KnownResourceReference{},
			Properties: batch.BatchAccountCreateProperties{},
		},
	}
}

func Test_NewStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	account := createDummyResource()

	status, err := NewEmptyStatus(account)
	g.Expect(err).To(BeNil())
	g.Expect(status).To(BeAssignableToTypeOf(&batch.BatchAccount_Status{}))
}

func Test_EmptyArmResourceStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	account := createDummyResource()

	status, err := NewEmptyArmResourceStatus(account)
	g.Expect(err).To(BeNil())
	g.Expect(status).To(BeAssignableToTypeOf(&batch.BatchAccount_StatusArm{}))
}


func Test_HasStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	account := createDummyResource()
	result, err := HasStatus(account)
	g.Expect(err).To(BeNil())
	g.Expect(result).To(BeFalse())
}

type DummyStruct struct{}

func Test_NewPtrFromStruct_ReturnsPtr(t *testing.T) {
	g := NewGomegaWithT(t)

	v := DummyStruct{}
	ptr := NewPtrFromValue(v)
	g.Expect(ptr).To(BeAssignableToTypeOf(&DummyStruct{}))
}

func Test_NewPtrFromPrimitive_ReturnsPtr(t *testing.T) {
	g := NewGomegaWithT(t)

	v := 5
	ptr := NewPtrFromValue(v)

	expectedValue := &v

	g.Expect(ptr).To(Equal(expectedValue))
}

func Test_NewPtrFromPtr_ReturnsPtrPtr(t *testing.T) {
	g := NewGomegaWithT(t)

	ptr := &DummyStruct{}
	ptrPtr := NewPtrFromValue(ptr)
	g.Expect(ptrPtr).To(BeAssignableToTypeOf(&ptr))
}
