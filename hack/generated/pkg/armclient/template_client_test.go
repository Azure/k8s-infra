/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package armclient

import (
	"context"
	"fmt"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"testing"
	"time"
)

// TODO: This type graph is temporary and was snipped from some generated code...

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
//Generated from: https://schema.management.azure.com/schemas/2017-09-01/Microsoft.Batch.json#/resourceDefinitions/batchAccounts
type BatchAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BatchAccountsSpec `json:"spec,omitempty"`
}

type BatchAccountsSpec struct {

	// +kubebuilder:validation:Required
	ApiVersion BatchAccountsSpecApiVersion `json:"apiVersion"`

	//AzureName: The name of the resource in Azure. This is often the same as the name
	//of the resource in Kubernetes but it doesn't have to be.
	AzureName string `json:"azureName"`

	// +kubebuilder:validation:Required
	//Location: The region in which to create the account.
	Location string `json:"location"`

	// +kubebuilder:validation:Required
	Owner genruntime.KnownResourceReference `group:"microsoft.resources" json:"owner" kind:"ResourceGroup"`

	// +kubebuilder:validation:Required
	//Properties: The properties of a Batch account.
	Properties BatchAccountCreateProperties `json:"properties"`

	//Tags: The user-specified tags associated with the account.
	Tags map[string]string `json:"tags"`
}

var _ genruntime.ArmTransformer = &BatchAccountsSpec{}

// FromArm converts from an Azure ARM object to a Kubernetes CRD object
func (batchAccountsSpec *BatchAccountsSpec) FromArm(owner genruntime.KnownResourceReference, armInput interface{}) error {
	typedInput, ok := armInput.(*BatchAccountsSpecArm)
	if !ok {
		return fmt.Errorf("unexpected type supplied for FromArm function. Expected BatchAccountsSpecArm, got %T", armInput)
	}
	batchAccountsSpec.ApiVersion = typedInput.ApiVersion
	batchAccountsSpec.AzureName = genruntime.ExtractKubernetesResourceNameFromArmName(typedInput.Name)
	batchAccountsSpec.Location = typedInput.Location
	batchAccountsSpec.Owner = owner
	var err error
	properties := BatchAccountCreateProperties{}
	err = properties.FromArm(owner, typedInput.Properties)
	if err != nil {
		return err
	}
	batchAccountsSpec.Properties = properties
	batchAccountsSpec.Tags = typedInput.Tags
	return nil
}

// ToArm converts from a Kubernetes CRD object to an ARM object
func (batchAccountsSpec *BatchAccountsSpec) ToArm(name string) (interface{}, error) {
	if batchAccountsSpec == nil {
		return nil, nil
	}
	result := BatchAccountsSpecArm{}
	result.ApiVersion = batchAccountsSpec.ApiVersion
	result.Location = batchAccountsSpec.Location
	result.Name = name // TODO: Changed!
	properties, err := batchAccountsSpec.Properties.ToArm(name)
	if err != nil {
		return nil, err
	}
	result.Properties = properties.(BatchAccountCreatePropertiesArm)
	result.Tags = batchAccountsSpec.Tags
	result.Type = BatchAccountsSpecTypeMicrosoftBatchBatchAccounts
	return &result, nil
}

//Generated from: https://schema.management.azure.com/schemas/2017-09-01/Microsoft.Batch.json#/definitions/BatchAccountCreateProperties
type BatchAccountCreateProperties struct {

	//AutoStorage: The properties related to the auto-storage account.
	AutoStorage *AutoStorageBaseProperties `json:"autoStorage,omitempty"`

	//KeyVaultReference: Identifies the Azure key vault associated with a Batch
	//account.
	KeyVaultReference *KeyVaultReference `json:"keyVaultReference,omitempty"`

	//PoolAllocationMode: The pool allocation mode also affects how clients may
	//authenticate to the Batch Service API. If the mode is BatchService, clients may
	//authenticate using access keys or Azure Active Directory. If the mode is
	//UserSubscription, clients must use Azure Active Directory. The default is
	//BatchService.
	PoolAllocationMode *BatchAccountCreatePropertiesPoolAllocationMode `json:"poolAllocationMode,omitempty"`
}

var _ genruntime.ArmTransformer = &BatchAccountCreateProperties{}

// FromArm converts from an Azure ARM object to a Kubernetes CRD object
func (batchAccountCreateProperties *BatchAccountCreateProperties) FromArm(owner genruntime.KnownResourceReference, armInput interface{}) error {
	typedInput, ok := armInput.(BatchAccountCreatePropertiesArm)
	if !ok {
		return fmt.Errorf("unexpected type supplied for FromArm function. Expected BatchAccountCreatePropertiesArm, got %T", armInput)
	}
	var err error
	if typedInput.AutoStorage != nil {
		autoStorage := AutoStorageBaseProperties{}
		err = autoStorage.FromArm(owner, typedInput.AutoStorage)
		if err != nil {
			return err
		}
		batchAccountCreateProperties.AutoStorage = &autoStorage
	}
	if typedInput.KeyVaultReference != nil {
		keyVaultReference := KeyVaultReference{}
		err = keyVaultReference.FromArm(owner, typedInput.KeyVaultReference)
		if err != nil {
			return err
		}
		batchAccountCreateProperties.KeyVaultReference = &keyVaultReference
	}
	batchAccountCreateProperties.PoolAllocationMode = typedInput.PoolAllocationMode
	return nil
}

// ToArm converts from a Kubernetes CRD object to an ARM object
func (batchAccountCreateProperties *BatchAccountCreateProperties) ToArm(owningName string) (interface{}, error) {
	if batchAccountCreateProperties == nil {
		return nil, nil
	}
	result := BatchAccountCreatePropertiesArm{}
	if batchAccountCreateProperties.AutoStorage != nil {
		autoStorage, err := batchAccountCreateProperties.AutoStorage.ToArm(owningName)
		if err != nil {
			return nil, err
		}
		autoStorageTyped := autoStorage.(AutoStorageBasePropertiesArm)
		result.AutoStorage = &autoStorageTyped
	}
	if batchAccountCreateProperties.KeyVaultReference != nil {
		keyVaultReference, err := batchAccountCreateProperties.KeyVaultReference.ToArm(owningName)
		if err != nil {
			return nil, err
		}
		keyVaultReferenceTyped := keyVaultReference.(KeyVaultReferenceArm)
		result.KeyVaultReference = &keyVaultReferenceTyped
	}
	result.PoolAllocationMode = batchAccountCreateProperties.PoolAllocationMode
	return result, nil
}

//Generated from: https://schema.management.azure.com/schemas/2017-09-01/Microsoft.Batch.json#/definitions/KeyVaultReference
type KeyVaultReference struct {

	// +kubebuilder:validation:Required
	//Id: The resource ID of the Azure key vault associated with the Batch account.
	Id string `json:"id"`

	// +kubebuilder:validation:Required
	//Url: The URL of the Azure key vault associated with the Batch account.
	Url string `json:"url"`
}

var _ genruntime.ArmTransformer = &KeyVaultReference{}

// FromArm converts from an Azure ARM object to a Kubernetes CRD object
func (keyVaultReference *KeyVaultReference) FromArm(owner genruntime.KnownResourceReference, armInput interface{}) error {
	typedInput, ok := armInput.(KeyVaultReferenceArm)
	if !ok {
		return fmt.Errorf("unexpected type supplied for FromArm function. Expected KeyVaultReferenceArm, got %T", armInput)
	}
	keyVaultReference.Id = typedInput.Id
	keyVaultReference.Url = typedInput.Url
	return nil
}

// ToArm converts from a Kubernetes CRD object to an ARM object
func (keyVaultReference *KeyVaultReference) ToArm(owningName string) (interface{}, error) {
	if keyVaultReference == nil {
		return nil, nil
	}
	result := KeyVaultReferenceArm{}
	result.Id = keyVaultReference.Id
	result.Url = keyVaultReference.Url
	return result, nil
}

type BatchAccountsSpecArm struct {
	ApiVersion BatchAccountsSpecApiVersion `json:"apiVersion"`

	//Location: The region in which to create the account.
	Location string `json:"location"`

	//Name: A name for the Batch account which must be unique within the region. Batch
	//account names must be between 3 and 24 characters in length and must use only
	//numbers and lowercase letters. This name is used as part of the DNS name that is
	//used to access the Batch service in the region in which the account is created.
	//For example: http://accountname.region.batch.azure.com/.
	Name string `json:"name"`

	//Properties: The properties of a Batch account.
	Properties BatchAccountCreatePropertiesArm `json:"properties"`

	//Tags: The user-specified tags associated with the account.
	Tags map[string]string     `json:"tags"`
	Type BatchAccountsSpecType `json:"type"`
}

var _ genruntime.ArmResourceSpec = &BatchAccountsSpecArm{}

func (spec *BatchAccountsSpecArm) AzureApiVersion() string {
	return string(spec.ApiVersion)
}

// TODO: this seems a bit wonky
func (spec *BatchAccountsSpecArm) ResourceGroup() string {
	return strings.Split(spec.Name, "/")[0]
}

//Generated from: https://schema.management.azure.com/schemas/2017-09-01/Microsoft.Batch.json#/definitions/BatchAccountCreateProperties
type BatchAccountCreatePropertiesArm struct {

	//AutoStorage: The properties related to the auto-storage account.
	AutoStorage *AutoStorageBasePropertiesArm `json:"autoStorage,omitempty"`

	//KeyVaultReference: Identifies the Azure key vault associated with a Batch
	//account.
	KeyVaultReference *KeyVaultReferenceArm `json:"keyVaultReference,omitempty"`

	//PoolAllocationMode: The pool allocation mode also affects how clients may
	//authenticate to the Batch Service API. If the mode is BatchService, clients may
	//authenticate using access keys or Azure Active Directory. If the mode is
	//UserSubscription, clients must use Azure Active Directory. The default is
	//BatchService.
	PoolAllocationMode *BatchAccountCreatePropertiesPoolAllocationMode `json:"poolAllocationMode,omitempty"`
}

//Generated from: https://schema.management.azure.com/schemas/2017-09-01/Microsoft.Batch.json#/definitions/AutoStorageBaseProperties
type AutoStorageBaseProperties struct {

	// +kubebuilder:validation:Required
	//StorageAccountId: The resource ID of the storage account to be used for
	//auto-storage account.
	StorageAccountId string `json:"storageAccountId"`
}

var _ genruntime.ArmTransformer = &AutoStorageBaseProperties{}

// FromArm converts from an Azure ARM object to a Kubernetes CRD object
func (autoStorageBaseProperties *AutoStorageBaseProperties) FromArm(owner genruntime.KnownResourceReference, armInput interface{}) error {
	typedInput, ok := armInput.(AutoStorageBasePropertiesArm)
	if !ok {
		return fmt.Errorf("unexpected type supplied for FromArm function. Expected AutoStorageBasePropertiesArm, got %T", armInput)
	}
	autoStorageBaseProperties.StorageAccountId = typedInput.StorageAccountId
	return nil
}

// ToArm converts from a Kubernetes CRD object to an ARM object
func (autoStorageBaseProperties *AutoStorageBaseProperties) ToArm(owningName string) (interface{}, error) {
	if autoStorageBaseProperties == nil {
		return nil, nil
	}
	result := AutoStorageBasePropertiesArm{}
	result.StorageAccountId = autoStorageBaseProperties.StorageAccountId
	return result, nil
}

//Generated from: https://schema.management.azure.com/schemas/2017-09-01/Microsoft.Batch.json#/definitions/AutoStorageBaseProperties
type AutoStorageBasePropertiesArm struct {

	//StorageAccountId: The resource ID of the storage account to be used for
	//auto-storage account.
	StorageAccountId string `json:"storageAccountId"`
}


//Generated from: https://schema.management.azure.com/schemas/2017-09-01/Microsoft.Batch.json#/definitions/KeyVaultReference
type KeyVaultReferenceArm struct {

	//Id: The resource ID of the Azure key vault associated with the Batch account.
	Id string `json:"id"`

	//Url: The URL of the Azure key vault associated with the Batch account.
	Url string `json:"url"`
}


// +kubebuilder:validation:Enum={"2017-09-01"}
type BatchAccountsSpecApiVersion string

const BatchAccountsSpecApiVersion20170901 = BatchAccountsSpecApiVersion("2017-09-01")

// +kubebuilder:validation:Enum={"Microsoft.Batch/batchAccounts"}
type BatchAccountsSpecType string

const BatchAccountsSpecTypeMicrosoftBatchBatchAccounts = BatchAccountsSpecType("Microsoft.Batch/batchAccounts")

// +kubebuilder:validation:Enum={"BatchService","UserSubscription"}
type BatchAccountCreatePropertiesPoolAllocationMode string

const (
	BatchAccountCreatePropertiesPoolAllocationModeBatchService     = BatchAccountCreatePropertiesPoolAllocationMode("BatchService")
	BatchAccountCreatePropertiesPoolAllocationModeUserSubscription = BatchAccountCreatePropertiesPoolAllocationMode("UserSubscription")
)

// TODO: This is duplicate code and should be deleted
func CreateDeploymentName() (string, error) {
	// no status yet, so start provisioning
	deploymentUUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	deploymentName := fmt.Sprintf("%s_%d_%s", "k8s", time.Now().Unix(), deploymentUUID.String())
	return deploymentName, nil
}

func Test_AllocateBatchAccountToAzure(t *testing.T) {
	g := NewGomegaWithT(t)

	armApplier, err := NewAzureTemplateClient()
	if err != nil {
		t.Fatal(err)
	}

	// Make a resource
	batchAccount := BatchAccount{
		Spec: BatchAccountsSpec{
			Owner: genruntime.KnownResourceReference{
				Name: "matthchr-rg",
			},
			Location: "westcentralus",
			ApiVersion: BatchAccountsSpecApiVersion20170901,
			AzureName: "matthchrk8s2",
		},
	}

	// For now just hardcode a resource group name
	rgName := "matthchr-rg"
	armObj, err := batchAccount.Spec.ToArm(batchAccount.Spec.AzureName)
	g.Expect(err).To(BeNil())

	typedArmObj, ok := armObj.(genruntime.ArmResourceSpec)
	g.Expect(ok).To(BeTrue())

	deploymentName, err := CreateDeploymentName()
	g.Expect(err).To(BeNil())
	deployment := armApplier.NewDeployment(rgName, deploymentName, typedArmObj)

	// Should be able to call apply a number of times and it works
	for !deployment.IsTerminalProvisioningState() {
		deployment, err = armApplier.ApplyDeployment(context.TODO(), deployment)
		g.Expect(err).To(BeNil())

		print(fmt.Sprintf("ProvState: %s\n", deployment.ProvisioningStateOrUnknown()))
		time.Sleep(2 * time.Second)
	}
}