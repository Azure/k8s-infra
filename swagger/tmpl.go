package swagger

//TODO: generate the types and the kubebuilder annotations
const Tmpl = `
{{ define "Path" -}}{{if .Operations.Get.Permitted}}
{ 
	Display: "{{ .Name}}",
	Endpoint: swagger.MustGetEndpointInfoFromURL("{{ .Operations.Get.Endpoint.TemplateURL }}", "{{ .Operations.Get.Endpoint.APIVersion}}"),
	{{- if ne .Operations.Get.Verb "" }}
	Verb: "{{upper .Operations.Get.Verb }}",{{end}}
	{{- if .Operations.Delete.Permitted }}
	DeleteEndpoint: swagger.MustGetEndpointInfoFromURL("{{ .Operations.Delete.Endpoint.TemplateURL }}", "{{ .Operations.Delete.Endpoint.APIVersion}}"),{{end}}
	{{- if .Operations.Patch.Permitted }}
	PatchEndpoint: swagger.MustGetEndpointInfoFromURL("{{ .Operations.Patch.Endpoint.TemplateURL }}", "{{ .Operations.Patch.Endpoint.APIVersion}}"),{{end}}
	{{- if .Operations.Put.Permitted }}
	PutEndpoint: swagger.MustGetEndpointInfoFromURL("{{ .Operations.Put.Endpoint.TemplateURL }}", "{{ .Operations.Put.Endpoint.APIVersion}}"),{{end}}
	{{- if .Children}}
	Children: {{template "PathList" .Children}},{{end}}
	{{- if .SubPaths}}
	SubResources: {{template "PathList" .SubPaths}},{{end}}
},{{end }}{{ end }}
{{define "PathList"}}[]swagger.ResourceType{ {{range .}}{{template "Path" .}}{{end}} } {{end}}
package expanders
import (
	"github.com/Azure/k8s-infra/swagger"
)
func (e *{{ .StructName }}) loadResourceTypes() []swagger.ResourceType {
	return  {{template "PathList" .Paths }}
}
`
