package jsonast

import (
	"context"
	"go/ast"
	"reflect"
	"testing"

	"github.com/xeipuuv/gojsonschema"

	. "github.com/onsi/gomega"
)

func TestToNodes(t *testing.T) {
	type args struct {
		resourcesSchema *gojsonschema.SubSchema
		opts            []BuilderOption
	}
	tests := []struct {
		name        string
		argsFactory func(*testing.T) *args
		want        []*ast.Package
		wantErr     bool
	}{
		{
			name: "WithSchema",
			argsFactory: func(t *testing.T) *args {
				schema, err := getDefaultSchema()
				if err != nil {
					t.Error(err)
				}

				return &args{
					resourcesSchema: schema,
				}
			},
			want:    nil,
			wantErr: false,
		},
	}
	scanner := NewSchemaScanner()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg := tt.argsFactory(t)
			got, err := scanner.ToNodes(context.TODO(), arg.resourcesSchema.ItemsChildren[0], arg.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToNodes() got = %+v, want %v", got, tt.want)
			}
		})
	}
}

func TestObjectWithNoType(t *testing.T) {
	schema := `
{
  "definitions": {
    "ApplicationSecurityGroupPropertiesFormat": {
      "description": "Application security group properties."
    }
  },

  "type": "object",
  "properties": {
    "foo": {
      "$ref": "#/definitions/ApplicationSecurityGroupPropertiesFormat"
    }
  }
}
`
	scanner := NewSchemaScanner()
	g := NewGomegaWithT(t)
	sl := gojsonschema.NewSchemaLoader()
	loader := gojsonschema.NewBytesLoader([]byte(schema))
	sb, err := sl.Compile(loader)
	g.Expect(err).To(BeNil())

	nodes, err := scanner.ToNodes(context.TODO(), sb.Root())

	g.Expect(err).To(BeNil())
	g.Expect(nodes).To(HaveLen(1))
	g.Expect(scanner.Structs).To(HaveLen(1))
	structDefinition := scanner.Structs[0]
	g.Expect(structDefinition.FieldCount()).To(Equal(1))
	propertiesField := structDefinition.Field(0)
	g.Expect(propertiesField.Name()).To(Equal("foo"))
	g.Expect(propertiesField.FieldType()).To(Equal("interface{}"))
}

/*
func XTestAnyOfWithMultipleComplexObjects(t *testing.T) {
	schema := `
{
  "definitions": {
    "genericExtension": {
      "type": "object",
      "properties": {
        "publisher": {
          "type": "string",
          "minLength": 1,
          "description": "Microsoft.Compute/extensions - Publisher"
        }
      }
    },
    "iaaSDiagnostics": {
      "type": "object",
      "properties": {
        "publisher": {
          "enum": [
            "Microsoft.Azure.Diagnostics"
          ]
        }
      }
    }
  },

  "type": "object",
  "properties": {
	"properties": {
	  "anyOf": [
		{
		  "$ref": "#/definitions/genericExtension"
		},
		{
		  "$ref": "#/definitions/iaaSDiagnostics"
		}
	  ]
	}
  },
  "description": "Microsoft.Compute/virtualMachines/extensions"
}
`
	scanner := &SchemaScanner{}
	g := NewGomegaWithT(t)
	sl := gojsonschema.NewSchemaLoader()
	loader := gojsonschema.NewBytesLoader([]byte(schema))
	sb, err := sl.Compile(loader)
	g.Expect(err).To(BeNil())
	nodes, err := scanner.ToNodes(context.TODO(), sb.Root())
	g.Expect(err).To(BeNil())
	g.Expect(nodes).To(HaveLen(1))
	structType, ok := nodes[0].(*ast.StructType)
	g.Expect(ok).To(BeTrue())
	g.Expect(structType.Fields.List).To(HaveLen(1))
	propertiesField := structType.Fields.List[0]
	g.Expect(propertiesField.Names[0]).To(Equal(ast.NewIdent("properties")))
	g.Expect(propertiesField.Type).To(Equal(ast.NewIdent("interface{}")))
}

func TestOneOfWithPropertySibling(t *testing.T) {
	schema := `
{
  "def": {
    "type": "object",
    "oneOf": [
      {
        "properties": {
          "ruleSetType": {
            "oneOf": [
              {
                "type": "string",
                "enum": [
                  "AzureManagedRuleSet"
                ]
              },
              {
                "$ref": "https://schema.management.azure.com/schemas/common/definitions.json#/definitions/expression"
              }
            ]
          }
        }
      }
    ],
    "properties": {
      "ruleSetType": {
        "type": "string"
      }
    },
    "required": [
      "ruleSetType"
    ],
    "description": "Describes azure managed provider."
  },
  "type": "object",
  "properties": {
    "ruleSets": {
      "oneOf": [
        {
          "type": "array",
          "items": {
            "$ref": "#/def"
          }
        },
        {
          "$ref": "https://schema.management.azure.com/schemas/common/definitions.json#/definitions/expression"
        }
      ],
      "description": "List of rules"
    }
  },
  "description": "Defines ManagedRuleSets - array of managedRuleSet"
}
`
	scanner := &SchemaScanner{}
	g := NewGomegaWithT(t)
	sl := gojsonschema.NewSchemaLoader()
	loader := gojsonschema.NewBytesLoader([]byte(schema))
	sb, err := sl.Compile(loader)
	g.Expect(err).To(BeNil())
	nodes, err := scanner.ToNodes(context.TODO(), sb.Root())
	g.Expect(err).To(BeNil())
	g.Expect(nodes).To(HaveLen(1))
	structType, ok := nodes[0].(*ast.StructType)
	g.Expect(ok).To(BeTrue())
	g.Expect(structType.Fields.List).To(HaveLen(1))
}

func TestAllOfUnion(t *testing.T) {
	schema := `{
  "definitions": {
    "address": {
      "type": "object",
      "properties": {
        "street_address": { "type": "string" },
        "city":           { "type": "string" },
        "state":          { "type": "string" }
      },
      "required": ["street_address", "city", "state"]
    }
  },

  "allOf": [
    { "$ref": "#/definitions/address" },
    { "properties": {
        "type": { "enum": [ "residential", "business" ] }
      }
    }
  ]
}`
	scanner := &SchemaScanner{}
	g := NewGomegaWithT(t)
	sl := gojsonschema.NewSchemaLoader()
	loader := gojsonschema.NewBytesLoader([]byte(schema))
	sb, err := sl.Compile(loader)
	g.Expect(err).To(BeNil())
	nodes, err := scanner.ToNodes(context.TODO(), sb.Root())
	g.Expect(err).To(BeNil())
	g.Expect(nodes).To(HaveLen(1))
	structType, ok := nodes[0].(*ast.StructType)
	g.Expect(ok).To(BeTrue())
	g.Expect(structType.Fields.List).To(HaveLen(4))
}

func TestAnyOfLocation(t *testing.T) {
	schema := `
{
"anyOf": [
	{
	  "type": "string"
	},
	{
	  "enum": [
		"East Asia",
		"Southeast Asia",
		"Central US"
	  ]
	}
  ]
}`
	scanner := &SchemaScanner{}
	g := NewGomegaWithT(t)
	sl := gojsonschema.NewSchemaLoader()
	loader := gojsonschema.NewBytesLoader([]byte(schema))
	sb, err := sl.Compile(loader)
	g.Expect(err).To(BeNil())
	nodes, err := scanner.ToNodes(context.TODO(), sb.Root())
	g.Expect(err).To(BeNil())
	g.Expect(nodes).To(HaveLen(1))
	field, ok := nodes[0].(*ast.Field)
	g.Expect(ok).To(BeTrue())
	g.Expect(field.Names[0].Name).To(Equal("anyOf"))
	g.Expect(field.Type.(*ast.Ident)).To(Equal(ast.NewIdent("string")))
}

func getDefaultSchema() (*gojsonschema.SubSchema, error) {
	sl := gojsonschema.NewSchemaLoader()
	ref := "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json"
	schema, err := sl.Compile(gojsonschema.NewReferenceLoader(ref))
	if err != nil {
		return nil, err
	}

	root := schema.Root()
	for _, child := range root.PropertiesChildren {
		if child.Property == "resources" {
			return child, nil
		}
	}
	return nil, errors.New("couldn't find resources in the schema")
}
*/
