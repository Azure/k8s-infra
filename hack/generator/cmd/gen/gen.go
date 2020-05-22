/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package gen

import (
	"context"
	"github.com/Azure/k8s-infra/hack/generator/pkg/codegen"
	"github.com/Azure/k8s-infra/hack/generator/pkg/xcobra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"log"
)

// NewGenCommand creates a new cobra Command when invoked from the command line
func NewGenCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "gen",
		Short: "generate K8s infrastructure resources from Azure deployment template schema",
		Run: xcobra.RunWithCtx(func(ctx context.Context, cmd *cobra.Command, args []string) error {

			cg, err := codegen.NewCodeGenerator("azure-cloud.yaml")
			if err != nil {
				log.Printf("Error: %v\n", err)
				return err
			}

			err = cg.Generate(ctx, "apis")
			if err != nil {
				klog.Errorf("Error during code generation:\n%v\n", err)
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
