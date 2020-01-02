package main

// kubebuilder validation annotations and description comment come from the swagger
//+kubebuilder:validation:Required
//+kubebuilder:validation:MinLength=1
//+kubebuilder:validation:MaxLength=90
//+kubebuilder:validation:Pattern=^[-\\w\\._\\(\\)]+$
// The name of the resource group.
// Name  string `validation-type:"create-only"`
// Location string `validation-type:"create-only"`
// ID string `validation-type:"read-only"`

const tmpl = `
package api

type ResourceGroupSpec struct {
	ID string {{.Backtick}}pathTemplate="{{.IdTemplate}}"{{.Backtick}}
	// Create-only props
	{{range $name, $spec := .WriteOnceProps }} 
	{{- specField $name}} 
	{{- $name}} {{type $name $spec}}
	{{end}}
	// Normal props
	{{range $name, $spec := .NormalProps }} 
	{{- specField $name}} 
	{{- $name}} {{type $name $spec}} 
	{{end}}
}

type ResourceGroupStatus struct {
	{{range $name, $spec := .ReadonlyProps }} 
	{{- $name}} {{type $name $spec}} 
	{{end}}
 }
`
