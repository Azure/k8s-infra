package main

type ResourceGroupSpec struct {
	ID string `pathTemplate="/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}"`
	// Create-only props
	location string
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:Pattern=^[-\w\._\(\)]+$
	//+kubebuilder:validation:MaxLength=90
	//+kubebuilder:validation:MinLength=1
	resourceGroupName string
	//+kubebuilder:validation:Required
	subscriptionId string
	
	// Normal props
	managedBy string 
	tags map[string]*string 
	
}

type ResourceGroupStatus struct {
	id string 
	name string 
	provisioningState string 
	type string 
	
 }
