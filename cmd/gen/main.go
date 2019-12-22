package main

import (
	"fmt"
	"log"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

const (
	InURLPathParam = "path"
	InBodyParam    = "body"
)

var inPathParams = map[string]spec.Parameter{}
var writeOnceProps = map[string]*spec.Schema{}
var readonlyProps = map[string]*spec.Schema{}
var normalProps = map[string]*spec.Schema{}

func main() {

	// Todo: build from azure-rest-specs repo - see azbrowse

	doc := loadDoc("/workspace/specs/azure-rest-api-specs/specification/resources/resource-manager/Microsoft.Resources/stable/2019-05-01/resources.json")

	// Todo: determine root object of interest - see azbrowse

	pathItem := doc.Spec().Paths.Paths["/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}"]

	putParams := pathItem.Put.Parameters
	for _, param := range putParams {
		if param.In == InURLPathParam {
			writeOnceProps[param.Name] = param.Schema
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
}

func mapPropertiesToType(properties map[string]spec.Schema, param spec.Parameter) {
	for bodyPropName, bodyPropSchema := range properties {
		fmt.Println(param.Name, bodyPropName, bodyPropSchema)

		// Relevant as means it can be set ... but only at create time.
		_, existsInAsPathParam := inPathParams[bodyPropName]
		if bodyPropSchema.ReadOnly {
			if existsInAsPathParam {
				// If it's readonly in the body but exists in path
				// then it's a create only property (can be read in body but only set at create)
				writeOnceProps[bodyPropName] = &bodyPropSchema
			} else {
				// readonly property
				readonlyProps[bodyPropName] = &bodyPropSchema
			}
		} else {
			if len(bodyPropSchema.Properties) > 0 {
				mapPropertiesToType(bodyPropSchema.Properties, param)
				continue
			}
			normalProps[bodyPropName] = &bodyPropSchema
		}

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
