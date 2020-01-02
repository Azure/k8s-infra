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

	pathItem := doc.Spec().Paths.Paths["/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}"]

	idTemplate := "/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}"
	backtick := "`"

	patchParams := pathItem.Patch.Parameters
	for _, param := range patchParams {
		if param.In == "body" {
			for bodyPropName, bodyPropSchema := range param.Schema.Properties {
				patchProps[bodyPropName] = bodyPropSchema
			}
		}
	}

	putParams := pathItem.Put.Parameters
	for _, param := range putParams {
		if param.In == InURLPathParam {
			writeOnceProps[param.Name] = spec.Schema{}
			inPathParams[param.Name] = param
		}

		if param.In == "body" {
			mapPropertiesToType(param.Schema.Properties, param)
		}
	}

	fmt.Printf("InPathParams: %+v \n", inPathParams)
	fmt.Printf("WriteOnce: %+v \n", writeOnceProps)
	fmt.Printf("Readonly: %+v \n", readonlyProps)
	fmt.Printf("Normal: %+v \n", normalProps)

	kubebuilderCommentSpec := func(name string) string {
		builder := strings.Builder{}
		// Todo: hack

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

	funcMap := template.FuncMap{
		"specField":   kubebuilderCommentSpec,
		"statusField": kubebuilderCommentSpec,
		"type":        typeFunc,
	}
	// Create a template, add the function map, and parse the text.
	tmpl, err := template.New("crd").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		log.Fatalf("parsing: %s", err)
	}

	// Run the template to verify the output.
	err = tmpl.Execute(os.Stdout, templateInput{
		WriteOnceProps: writeOnceProps,
		ReadonlyProps:  readonlyProps,
		NormalProps:    normalProps,
		IdTemplate:     idTemplate,
		Backtick:       backtick,
	})
	if err != nil {
		log.Fatalf("execution: %s", err)
	}
}

func mapPropertiesToType(properties map[string]spec.Schema, param spec.Parameter) {
	for bodyPropName, bodyPropSchema := range properties {
		fmt.Println(param.Name, bodyPropName, bodyPropSchema)

		// Relevant as means it can be set ... but only at create time.
		pathSchema, existsInAsPathParam := inPathParams[bodyPropName]
		if bodyPropSchema.ReadOnly {
			if existsInAsPathParam {
				// If it's readonly in the body but exists in path
				// then it's a create only property (can be read in body but only set at create)
				writeOnceProps[bodyPropName] = *pathSchema.Schema
				continue
			} else {
				// readonly property
				readonlyProps[bodyPropName] = bodyPropSchema
				continue
			}
		}

		if len(bodyPropSchema.Properties) > 0 {
			mapPropertiesToType(bodyPropSchema.Properties, param)
			continue
		}

		_, existsInPatchBody := patchProps[bodyPropName]
		// If the body param doesn't exist in the patch body it's create only
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
