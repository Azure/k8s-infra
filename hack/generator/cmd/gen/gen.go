package gen

import (
	"context"
	"fmt"
	"os"
	"unicode"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xeipuuv/gojsonschema"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
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

			//schema, err := loadSchema(rgTemplateSchemaFile2)
			schema, err := loadSchema(rgTemplateSchemaURI)
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

			scanner := jsonast.NewSchemaScanner()
			scanner.AddFilters(viper.GetStringSlice("resources"))
			_, err = scanner.ToNodes(ctx, resourcesSchema)
			if err != nil {
				fmt.Printf("GEN0048 - Error %s\n", err)
				return err
			}

			err = os.RemoveAll("resources")
			if err != nil {
				fmt.Printf("GEN0054 - Error %s\n", err)
				return err
			}

			err = os.Mkdir("resources", 0700)
			if err != nil {
				fmt.Printf("GEN0060 - Error %s\n", err)
				return err
			}

			fmt.Printf("GEN0064 INF Checkpoint\n")

			for _, st := range scanner.Structs {
				ns := createNamespace("api", st.Version())
				dirName := fmt.Sprintf("resources/%v", ns)
				fileName := fmt.Sprintf("%v/%v.go", dirName, st.Name())

				if _, err := os.Stat(dirName); os.IsNotExist(err) {
					fmt.Printf("GEN072 - Creating folder %s\n", dirName)
					os.Mkdir(dirName, 0700)
				}

				fmt.Printf("GEN076 - Writing %s\n", fileName)

				genFile := astmodel.NewFileDefinition(ns, st)
				genFile.SaveTo(fileName)
			}

			fmt.Printf("GEN061 - Completed creating resources\n")

			return nil
		}),
	}

	cmd.Flags().StringArrayP("resources", "r", nil, "list of resource type / versions to generate")
	if err := viper.BindPFlag("resources", cmd.Flags().Lookup("resources")); err != nil {
		return cmd, err
	}

	return cmd, nil
}

func loadSchema(source string) (*gojsonschema.Schema, error) {
	sl := gojsonschema.NewSchemaLoader()
	schema, err := sl.Compile(gojsonschema.NewReferenceLoader(source))
	if err != nil {
		return nil, err
	}

	return schema, nil
}

func createNamespace(base string, version string) string {
	var builder []rune

	for _, r := range base {
		builder = append(builder, rune(r))
	}

	for _, r := range version {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder = append(builder, rune(r))
		}
	}

	return string(builder)
}
