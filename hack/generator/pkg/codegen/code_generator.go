/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"
	"io/ioutil"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/Azure/k8s-infra/hack/generator/pkg/config"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

// CodeGenerator is a generator of code
type CodeGenerator struct {
	configuration *config.Configuration
	pipeline      []PipelineStage
}

// NewCodeGeneratorFromConfigFile produces a new Generator with the given configuration file
func NewCodeGeneratorFromConfigFile(configurationFile string) (*CodeGenerator, error) {
	configuration, err := loadConfiguration(configurationFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load configuration file %q", configurationFile)
	}

	err = configuration.Initialize()
	if err != nil {
		return nil, errors.Wrapf(err, "configuration loaded from %q is invalid", configurationFile)
	}

	idFactory := astmodel.NewIdentifierFactory()

	return NewCodeGeneratorFromConfig(configuration, idFactory)
}

// NewCodeGeneratorFromConfig produces a new Generator with the given configuration
func NewCodeGeneratorFromConfig(configuration *config.Configuration, idFactory astmodel.IdentifierFactory) (*CodeGenerator, error) {
	var pipeline []PipelineStage
	pipeline = append(pipeline, loadSchemaIntoTypes(idFactory, configuration, defaultSchemaLoader))
	pipeline = append(pipeline, corePipelineStages(idFactory, configuration)...)
	pipeline = append(pipeline, deleteGeneratedCode(configuration.OutputPath), exportPackages(configuration.OutputPath))

	result := &CodeGenerator{
		configuration: configuration,
		pipeline:      pipeline,
	}

	return result, nil
}

func corePipelineStages(idFactory astmodel.IdentifierFactory, configuration *config.Configuration) []PipelineStage {
	return []PipelineStage{
		nameTypesForCRD(idFactory),
		removeTypeAliases(),
		improveResourcePluralization(),
		applyExportFilters(configuration),
		stripUnreferencedTypeDefinitions(),
		checkForAnyType(),
		createStorageTypes(),
	}
}

// Generate produces the Go code corresponding to the configured JSON schema in the given output folder
func (generator *CodeGenerator) Generate(ctx context.Context) error {
	klog.V(1).Infof("Generator version: %v", combinedVersion())

	defs := make(astmodel.Types)
	var err error

	for i, stage := range generator.pipeline {
		klog.V(0).Infof("Pipeline stage %d/%d: %s", i+1, len(generator.pipeline), stage.description)
		defs, err = stage.Action(ctx, defs)
		if err != nil {
			return errors.Wrapf(err, "Failed during pipeline stage %d/%d: %s", i+1, len(generator.pipeline), stage.description)
		}
	}

	return nil
}

func loadConfiguration(configurationFile string) (*config.Configuration, error) {
	data, err := ioutil.ReadFile(configurationFile)
	if err != nil {
		return nil, err
	}

	result := config.NewConfiguration()

	err = yaml.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
