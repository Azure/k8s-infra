package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

const (
	InURLPathParam = "path"
	InBodyParam    = "body"
)

var inPathParams = map[string]spec.Parameter{}
var writeOnceProps = map[string]spec.Schema{}
var readonlyProps = map[string]spec.Schema{}
var normalProps = map[string]spec.Schema{}
var patchProps = map[string]spec.Schema{}

type templateInput struct {
	WriteOnceProps map[string]spec.Schema
	ReadonlyProps  map[string]spec.Schema
	NormalProps    map[string]spec.Schema
	IdTemplate     string
	Backtick       string
}

func main() {

	// Todo: build from azure-rest-specs repo - see azbrowse
	doc := loadDoc("/workspace/specs/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2019-05-01/resources.json")

	// Todo: determine root object of interest - see azbrowse
	idTemplate := "/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}"
	pathItem := doc.Spec().Paths.Paths[idTemplate]

	// Walk the patch body params
	patchParams := pathItem.Patch.Parameters
	for _, param := range patchParams {
		if param.In == "body" {
			for bodyPropName, bodyPropSchema := range param.Schema.Properties {
				patchProps[bodyPropName] = bodyPropSchema
			}
		}
	}

	// Walk the put params (body and url)
	putParams := pathItem.Put.Parameters
	for _, param := range putParams {
		if param.In == InURLPathParam {
			writeOnceProps[param.Name] = spec.Schema{}
			inPathParams[param.Name] = param
		}

		if param.In == "body" {
			// Attempt to infer details on type of field from spec
			mapPropertiesToType(param.Schema.Properties, param)
		}
	}

	fmt.Printf("InPathParams: %+v \n", inPathParams)
	fmt.Printf("WriteOnce: %+v \n", writeOnceProps)
	fmt.Printf("Readonly: %+v \n", readonlyProps)
	fmt.Printf("Normal: %+v \n", normalProps)

	kubebuilderCommentSpec := func(name string) string {
		builder := strings.Builder{}

		// Todo: hack only looking at inPath as haven't seen
		// any body props with validation logic on them ... sadly
		prop, existsInAsPathParam := inPathParams[name]
		if !existsInAsPathParam {
			return ""
		}

		builder.WriteString("//+kubebuilder:validation:Required\n	")

		if prop.Pattern != "" {
			builder.WriteString(fmt.Sprintf("//+kubebuilder:validation:Pattern=%s\n	", prop.Pattern))
		}

		if prop.MaxLength != nil {
			builder.WriteString(fmt.Sprintf("//+kubebuilder:validation:MaxLength=%v\n	", *prop.MaxLength))
		}

		if prop.MinLength != nil {
			builder.WriteString(fmt.Sprintf("//+kubebuilder:validation:MinLength=%v\n	", *prop.MinLength))
		}
		return builder.String()
	}

	typeFunc := func(name string, schema spec.Schema) string {
		if schema.Type.Contains("string") {
			return "string"
		}

		if schema.Type.Contains("object") {
			return "map[string]*string"
		}

		if schema.Type.Contains("array") {
			return "[]string"
		}

		if schema.Type.Contains("number") || schema.Type.Contains("integer") {
			return "int"
		}

		return "string"
	}

	jsonTagFunc := func(name string) string {
		return fmt.Sprintf(`json:"%s"`, name)
	}

	funcMap := template.FuncMap{
		"specField": kubebuilderCommentSpec,
		"type":      typeFunc,
		"jsonTag":   jsonTagFunc,
		"title":     strings.Title,
	}

	// Parse the template
	tmpl, err := template.New("crd").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		log.Fatalf("parsing: %s", err)
	}

	stringBuilder := strings.Builder{}

	// Remove ID field from props as it's hard coded in the template
	delete(readonlyProps, "Id")

	// Run the template to verify the output.
	err = tmpl.Execute(&stringBuilder, templateInput{
		WriteOnceProps: writeOnceProps,
		ReadonlyProps:  readonlyProps,
		NormalProps:    normalProps,
		IdTemplate:     idTemplate,
		Backtick:       "`",
	})
	if err != nil {
		log.Fatalf("execution: %s", err)
	}

	fmt.Print(stringBuilder.String())

	os.Remove("./model.go")
	f, err := os.Create("./model.go")
	defer f.Close()
	f.WriteString(stringBuilder.String())

}

func mapPropertiesToType(properties map[string]spec.Schema, param spec.Parameter) {
	for bodyPropName, bodyPropSchema := range properties {
		fmt.Println(param.Name, bodyPropName, bodyPropSchema)

		// What readonly properties do we have?
		if bodyPropSchema.ReadOnly {

			// If it's readonly in the body but exists in path it's create-only be definition
			// changing the url means different resource
			pathSchema, existsInAsPathParam := inPathParams[bodyPropName]
			if existsInAsPathParam {
				writeOnceProps[bodyPropName] = *pathSchema.Schema
				continue
			}

			// normal readonly property
			readonlyProps[bodyPropName] = bodyPropSchema
			continue
		}

		// Do we need to recurse into the property? (For example provisioningState is under properties.provisioningState)
		// Todo: How does serialisation handle flatting and unflattening these types need `json:properties.provisioningState` but don't think
		// that exists
		if len(bodyPropSchema.Properties) > 0 {
			mapPropertiesToType(bodyPropSchema.Properties, param)
			continue
		}

		// Hack - need to try more types and validate this in more detail
		// If the body param doesn't exist in the patch body it's create only
		_, existsInPatchBody := patchProps[bodyPropName]
		if !existsInPatchBody {
			writeOnceProps[bodyPropName] = bodyPropSchema
			continue
		}

		normalProps[bodyPropName] = bodyPropSchema

	}
}

func loadDoc(path string) *loads.Document {

	document, err := loads.Spec(path)
	if err != nil {
		log.Panicf("Error opening Swagger: %s", err)
	}

	document, err = document.Expanded(&spec.ExpandOptions{RelativeBase: path})
	if err != nil {
		log.Panicf("Error expanding Swagger: %s", err)
	}

	return document
}
