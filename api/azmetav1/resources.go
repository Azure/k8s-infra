package azmetav1

// TrackedResourceSpec defines the core meta properties of an Azure tracked resource
type TrackedResourceSpec struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Location string `json:"location,omitempty"`
}

// NestedResourceSpec defines the core meta properties of an Azure nested resource
type NestedResourceSpec struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
