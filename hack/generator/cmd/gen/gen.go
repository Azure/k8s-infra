package gen

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"

	"github.com/Azure/k8s-infra/hack/generator/pkg/jsonast"
	"github.com/Azure/k8s-infra/hack/generator/pkg/xcobra"
)

const (
	rgTemplateSchemaURI = "https://schema.management.azure.com/schemas/2015-01-01/deploymentTemplate.json"
)

func NewGenCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "generate K8s infrastructure resources from Azure deployment template schema",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {
			sl := gojsonschema.NewSchemaLoader()
			schema, err := sl.Compile(gojsonschema.NewReferenceLoader(rgTemplateSchemaURI))
			if err != nil {
				return err
			}

			root := schema.Root()
			var resourcesSchema *gojsonschema.SubSchema
			for _, child := range root.PropertiesChildren {
				if child.Property == "resources" {
					resourcesSchema = child
					break
				}
			}

			_, err = jsonast.ToNodes(ctx, resourcesSchema, jsonast.WithFilters(viper.GetStringSlice("resources")))

			if err != nil {
				fmt.Println(err)
				return err
			}
			return nil
		}),
	}

	cmd.Flags().StringArrayP("resources", "r", nil, "list of resource type / versions to generate")
	if err := viper.BindPFlag("resources", cmd.Flags().Lookup("resources")); err != nil {
		return cmd, err
	}

	return cmd, nil
}