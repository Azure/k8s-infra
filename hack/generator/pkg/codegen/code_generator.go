/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/Azure/k8s-infra/hack/generator/pkg/config"
)

// CodeGenerator is a generator of code
type CodeGenerator struct {
	configuration *config.Configuration
	pipeline      []PipelineStage
}

// NewCodeGeneratorFromConfigFile produces a new Generator with the given configuration file
func NewCodeGeneratorFromConfigFile(configurationFile string) (*CodeGenerator, error) {
	configuration, err := config.LoadConfiguration(configurationFile)
	if err != nil {
		return nil, err
	}

	return NewCodeGeneratorFromConfig(configuration, astmodel.NewIdentifierFactory())
}

// NewCodeGeneratorFromConfig produces a new Generator with the given configuration
func NewCodeGeneratorFromConfig(configuration *config.Configuration, idFactory astmodel.IdentifierFactory) (*CodeGenerator, error) {
	var pipeline []PipelineStage
	pipeline = append(pipeline, loadSchemaIntoTypes(idFactory, configuration, defaultSchemaLoader))
	pipeline = append(pipeline, crossplaneCorePipelineStages(idFactory, configuration)...)
	pipeline = append(pipeline, deleteGeneratedCode(configuration.OutputPath), exportPackages(configuration.OutputPath))

	result := &CodeGenerator{
		configuration: configuration,
		pipeline:      pipeline,
	}

	return result, nil
}

func corePipelineStages(idFactory astmodel.IdentifierFactory, configuration *config.Configuration) []PipelineStage {
	return []PipelineStage{
		// Import status info from Swagger:
		augmentResourcesWithStatus(idFactory, configuration),

		// Reduces oneOf/allOf types from schemas to object types:
		convertAllOfAndOneOfToObjects(idFactory),

		// Flatten out any nested resources created by allOf, etc. we want to do this before naming types or things
		// get named with names like Resource_Spec_Spec_Spec:
		flattenResources(), stripUnreferencedTypeDefinitions(),

		// Name all anonymous object and enum types (required by controller-gen):
		nameTypesForCRD(idFactory),

		// Apply property type rewrites from the config file
		// must come after nameTypesForCRD and convertAllOfAndOneOf so that objects are all expanded
		applyPropertyRewrites(configuration),

		// Figure out ARM resource owners:
		determineResourceOwnership(),

		// Strip out redundant type aliases:
		removeTypeAliases(),

		// De-pluralize resource types:
		improveResourcePluralization(),

		stripUnreferencedTypeDefinitions(),

		// Apply export filters before generating
		// ARM types for resources etc:
		applyExportFilters(configuration),
		stripUnreferencedTypeDefinitions(),
		filterOutDefinitionsUsingAnyType(configuration.AnyTypePackages),

		createArmTypesAndCleanKubernetesTypes(idFactory),
		applyKubernetesResourceInterface(idFactory),
		simplifyDefinitions(),

		// Safety checks at the end:
		ensureDefinitionsDoNotUseAnyTypes(),
		checkForMissingStatusInformation(),
	}
}

func crossplaneCorePipelineStages(idFactory astmodel.IdentifierFactory, configuration *config.Configuration) []PipelineStage {
	return []PipelineStage{
		// Import status info from Swagger:
		augmentResourcesWithStatus(idFactory, configuration),

		// Reduces oneOf/allOf types from schemas to object types:
		convertAllOfAndOneOfToObjects(idFactory),

		// Flatten out any nested resources created by allOf, etc. we want to do this before naming types or things
		// get named with names like Resource_Spec_Spec_Spec:
		flattenResources(), stripUnreferencedTypeDefinitions(),

		// Name all anonymous object and enum types (required by controller-gen):
		nameTypesForCRD(idFactory),

		// Apply property type rewrites from the config file
		// must come after nameTypesForCRD and convertAllOfAndOneOf so that objects are all expanded
		applyPropertyRewrites(configuration),

		// Figure out ARM resource owners:
		determineResourceOwnership(),

		// Strip out redundant type aliases:
		removeTypeAliases(),

		// De-pluralize resource types:
		// improveResourcePluralization(),

		stripUnreferencedTypeDefinitions(),

		// Apply export filters before generating
		// ARM types for resources etc:
		applyExportFilters(configuration),
		stripUnreferencedTypeDefinitions(),

		// filterOutDefinitionsUsingAnyType(configuration.AnyTypePackages),

		// createArmTypesAndCleanKubernetesTypes(idFactory),

		addForProviderProperty(idFactory),

		// applyKubernetesResourceInterface(idFactory),
		simplifyDefinitions(),

		// Safety checks at the end:
		// ensureDefinitionsDoNotUseAnyTypes(),
		checkForMissingStatusInformation(),
	}
}

// Generate produces the Go code corresponding to the configured JSON schema in the given output folder
func (generator *CodeGenerator) Generate(ctx context.Context) error {
	klog.V(1).Infof("Generator version: %v", combinedVersion())

	defs := make(astmodel.Types)
	for i, stage := range generator.pipeline {
		klog.V(0).Infof("Pipeline stage %d/%d: %s", i+1, len(generator.pipeline), stage.description)
		updatedDefs, err := stage.Action(ctx, defs)
		if err != nil {
			return errors.Wrapf(err, "Failed during pipeline stage %d/%d: %s", i+1, len(generator.pipeline), stage.description)
		}

		defsAdded := updatedDefs.Except(defs)
		defsRemoved := defs.Except(updatedDefs)

		if len(defsAdded) > 0 && len(defsRemoved) > 0 {
			klog.V(1).Infof("Added %d, removed %d type definitions", len(defsAdded), len(defsRemoved))
		} else if len(defsAdded) > 0 {
			klog.V(1).Infof("Added %d type definitions", len(defsAdded))
		} else if len(defsRemoved) > 0 {
			klog.V(1).Infof("Removed %d type definitions", len(defsRemoved))
		}

		defs = updatedDefs
	}

	return nil
}
