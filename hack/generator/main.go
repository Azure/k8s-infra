package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/xeipuuv/gojsonschema"
)

const (
	rgTemplateSchemaURI = "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json"
)

func main() {

	sl := gojsonschema.NewSchemaLoader()
	sl.Validate = false
	schema, err := sl.Compile(gojsonschema.NewReferenceLoader(rgTemplateSchemaURI))
	if err != nil {
		errAndExit(err)
	}

	/*
		The schema we are working with is something like the following (in yaml for brevity):
		properties:
			resources:
				items:
					oneOf:
						allOf:
							$ref: {{ base resource schema for ARM }}
							oneOf:
								- ARM resources
					oneOf:
						allOf:
							$ref: {{ base resource for external resources, think SendGrid }}
							oneOf:
								- External ARM resources
					oneOf:
						allOf:
							$ref: {{ base resource for ARM specific stuff like locks, deployments, etc }}
							oneOf:
								- ARM specific resources. I'm not 100% sure why...

		allOf acts like composition which composites each schema from the child oneOf with the base reference from allOf.
	 */
	root := schema.Root()
	var resourcesSchema *gojsonschema.SubSchema
	for _, child := range root.PropertiesChildren {
		if child.Property == "resources" {
			resourcesSchema = child
			break
		}
	}

	items := resourcesSchema.ItemsChildren[0]
	allOfAndOneOfs := items.OneOf[0].AllOf // where most of the types we are looking for reside
	propertiesFromBase := allOfAndOneOfs[0].RefSchema.AllOf[0].RefSchema.PropertiesChildren
	propertiesFromBase = append(propertiesFromBase, allOfAndOneOfs[0].RefSchema.AllOf[1].PropertiesChildren...)

	fmt.Println("####  BASE PROPERTIES SHARED BY ALL  ####")
	for _, propSchema := range propertiesFromBase {
		fmt.Println(propSchema.Property)
	}

	fmt.Println("\n\n####  Each Resource Type and API Version  ####")
	oneOfTheseResourceRefs := allOfAndOneOfs[1].OneOf
	typeAndVersion := make([]string, len(oneOfTheseResourceRefs))
	for i, resourceRef := range oneOfTheseResourceRefs {
		properties := resourceRef.RefSchema.PropertiesChildren
		var t, v string
		for _, prop := range properties {
			switch prop.Property {
			case "type":
				t = prop.Enum[0] // consts are shaped like single value enums
			case "apiVersion":
				v = prop.Enum[0] // see above
			}
		}
		typeAndVersion[i] = fmt.Sprintf("%s/%s", t[1:len(t)-1], v[1:len(v)-1]) // trim off the extra quotes
	}

	sort.Strings(typeAndVersion)
	for _, tv := range typeAndVersion {
		fmt.Println(tv)
	}
}

func errAndExit(err error) {
	fmt.Println(err)
	os.Exit(1)
}