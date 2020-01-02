package main

const tmpl = `
{{- $root := . -}}
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
	ID string {{.Backtick}}json:"id" pathTemplate:"{{.IdTemplate}}"{{.Backtick}}

	// Create-only props

	{{range $name, $spec := .WriteOnceProps }} 
	{{- specField $name}} 
	{{- title $name}} {{type $name $spec}} {{$root.Backtick}}validation-type:"create-only" {{jsonTag $name}}{{$root.Backtick}}
	{{end}}

	// Normal props

	{{range $name, $spec := .NormalProps }} 
	{{- specField $name}} 
	{{- title $name}} {{type $name $spec}} {{$root.Backtick}}validation-type:"normal" {{jsonTag $name}}{{$root.Backtick}}
	{{end}}
}

type ResourceGroupStatus struct {
	{{range $name, $spec := .ReadonlyProps }} 
	{{- title $name}} {{type $name $spec}} {{$root.Backtick}}validation-type:"read-only" {{jsonTag $name}}{{$root.Backtick}}
	{{end}}
 }
`
