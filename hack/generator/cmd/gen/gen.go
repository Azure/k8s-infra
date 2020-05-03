/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package gen

import (
	"context"
	"log"
	"os"
	"path/filepath"

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

// NewGenCommand creates a new cobra Command when invoked from the command line
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

			rootOutputDir := "apis" // TODO: command-line argument

			idfactory := astmodel.NewIdentifierFactory()
			scanner := jsonast.NewSchemaScanner(idfactory)
			scanner.AddFilters(viper.GetStringSlice("resources"))

			_, err = scanner.ToNodes(ctx, root)
			if err != nil {
				log.Printf("Error: %v\n", err)
				return err
			}

			err = os.RemoveAll(rootOutputDir)
			if err != nil {
				log.Printf("Error: %v\n", err)
				return err
			}

			err = os.Mkdir(rootOutputDir, 0700)
			if err != nil {
				log.Printf("Error %v\n", err)
				return err
			}

			log.Printf("INF Checkpoint\n")

			// group definitions by package
			packages := make(map[astmodel.PackageReference][]*astmodel.StructDefinition)
			for _, def := range scanner.Structs {
				packages[def.PackageReference] = append(packages[def.PackageReference], def)
			}

			// emit each package
			for p, packageDefs := range packages {

				// create directory if not already there
				outputDir := filepath.Join(rootOutputDir, p.PackagePath())
				if _, err := os.Stat(outputDir); os.IsNotExist(err) {
					log.Printf("Creating directory '%s'\n", outputDir)
					err = os.MkdirAll(outputDir, 0700)
					if err != nil {
						log.Fatalf("Unable to create directory '%s'", outputDir)
					}
				}

				// emit each definition
				for _, def := range packageDefs {
					genFile := astmodel.NewFileDefinition(def)
					outputFile := filepath.Join(outputDir, def.Name()+"_types.go")
					log.Printf("Writing '%s'\n", outputFile)
					genFile.SaveTo(outputFile)
				}
			}

			log.Printf("Completed writing %v resources\n", len(scanner.Structs))

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
