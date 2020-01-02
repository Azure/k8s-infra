package main

// ResourceGroup top level type allows easy deserialisation of response from Azure RP REST calls
// while still allowing K8s to track difference between Spec vs Status
type ResourceGroup struct {
	ResourceGroupSpec
	ResourceGroupStatus
}

func (t ResourceGroup) Status() ResourceGroupStatus {
	return t.ResourceGroupStatus
}

func (t ResourceGroup) Spec() ResourceGroupSpec {
	return t.ResourceGroupSpec
}

type ResourceGroupSpec struct {
	ID string `json:"id" pathTemplate:"/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}"`

	// Create-only props

	Location string `validation-type:"create-only" json:"location"`
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:Pattern=^[-\w\._\(\)]+$
	//+kubebuilder:validation:MaxLength=90
	//+kubebuilder:validation:MinLength=1
	ResourceGroupName string `validation-type:"create-only" json:"resourceGroupName"`
	//+kubebuilder:validation:Required
	SubscriptionId string `validation-type:"create-only" json:"subscriptionId"`

	// Normal props

	ManagedBy string             `validation-type:"normal" json:"managedBy"`
	Tags      map[string]*string `validation-type:"normal" json:"tags"`
}

type ResourceGroupStatus struct {
	Id                string `validation-type:"read-only" json:"id"`
	Name              string `validation-type:"read-only" json:"name"`
	ProvisioningState string `validation-type:"read-only" json:"provisioningState"`
	Type              string `validation-type:"read-only" json:"type"`
}
