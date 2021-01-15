/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package codegen

import (
	"context"

	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"
	"github.com/Azure/k8s-infra/hack/generator/pkg/config"
)

// CodeGenerator is a generator of code
type CodeGenerator struct {
	configuration *config.Configuration
	pipeline      []PipelineStage
}


func translatePipelineToTarget(pipeline config.GenerationPipeline) (PipelineTarget, error) {
	switch pipeline {
	case config.GenerationPipelineAzure:
		return ArmTarget, nil
	case config.GenerationPipelineCrossplane:
		return CrossplaneTarget, nil
	default:
		return PipelineTarget{}, errors.Errorf("unknown pipeline target kind %s", pipeline)
	}
}

// NewCodeGeneratorFromConfigFile produces a new Generator with the given configuration file
func NewCodeGeneratorFromConfigFile(configurationFile string) (*CodeGenerator, error) {
	configuration, err := config.LoadConfiguration(configurationFile)
	if err != nil {
		return nil, err
	}

	target, err := translatePipelineToTarget(configuration.Pipeline)
	if err != nil {
		return nil, err
	}

	return NewTargetedCodeGeneratorFromConfig(configuration, astmodel.NewIdentifierFactory(), target)
}

// NewTargetedCodeGeneratorFromConfig produces a new code generator with the given configuration and
// only the stages appropriate for the specfied target.
func NewTargetedCodeGeneratorFromConfig(
	configuration *config.Configuration, idFactory astmodel.IdentifierFactory, target PipelineTarget) (*CodeGenerator, error) {

	result, err := NewCodeGeneratorFromConfig(configuration, idFactory)
	if err != nil {
		return nil, errors.Wrapf(err, "creating pipeline targeting %v", target)
	}

	// Filter stages to use only those appropriate for our target
	var stages []PipelineStage
	for _, s := range result.pipeline {
		if s.IsUsedFor(target) {
			stages = append(stages, s)
		}
	}

	result.pipeline = stages

	err = result.verifyPipeline()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// NewCodeGeneratorFromConfig produces a new code generator with the given configuration all available stages
func NewCodeGeneratorFromConfig(configuration *config.Configuration, idFactory astmodel.IdentifierFactory) (*CodeGenerator, error) {
	result := &CodeGenerator{
		configuration: configuration,
		pipeline:      createAllPipelineStages(idFactory, configuration),
	}

	return result, nil
}

func createAllPipelineStages(idFactory astmodel.IdentifierFactory, configuration *config.Configuration) []PipelineStage {
	return []PipelineStage{

		loadSchemaIntoTypes(idFactory, configuration, defaultSchemaLoader),

		// Import status info from Swagger:
		augmentResourcesWithStatus(idFactory, configuration),

		// Reduces oneOf/allOf types from schemas to object types:
		convertAllOfAndOneOfToObjects(idFactory),

		// Flatten out any nested resources created by allOf, etc. we want to do this before naming types or things
		// get named with names like Resource_Spec_Spec_Spec:
		flattenResources(),

		stripUnreferencedTypeDefinitions(),

		// Name all anonymous object, enum, and validated types (required by controller-gen):
		nameTypesForCRD(idFactory),

		// Apply property type rewrites from the config file
		// Must come after nameTypesForCRD ('nameTypes)' and convertAllOfAndOneOfToObjects ('allof-anyof-objects') so
		// that objects are all expanded
		applyPropertyRewrites(configuration).
			RequiresPrerequisiteStages("nameTypes", "allof-anyof-objects"),

		// Figure out resource owners:
		determineResourceOwnership(),

		// Strip out redundant type aliases:
		removeTypeAliases(),

		// De-pluralize resource types
		// (Must come after type aliases are resolved)
		improveResourcePluralization().
			RequiresPrerequisiteStages("removeAliases"),

		stripUnreferencedTypeDefinitions(),

		// Apply export filters before generating
		// ARM types for resources etc:
		applyExportFilters(configuration),

		stripUnreferencedTypeDefinitions(),

		replaceAnyTypeWithJSON(),
		reportOnTypesAndVersions(configuration).UsedFor(ArmTarget), // TODO: For now only used for ARM

		createArmTypesAndCleanKubernetesTypes(idFactory).UsedFor(ArmTarget),
		applyKubernetesResourceInterface(idFactory).UsedFor(ArmTarget),

		addCrossplaneOwnerProperties(idFactory).UsedFor(CrossplaneTarget),
		addCrossplaneForProvider(idFactory).UsedFor(CrossplaneTarget),
		addCrossplaneAtProvider(idFactory).UsedFor(CrossplaneTarget),
		addCrossplaneEmbeddedResourceSpec(idFactory).UsedFor(CrossplaneTarget),
		addCrossplaneEmbeddedResourceStatus(idFactory).UsedFor(CrossplaneTarget),

		createStorageTypes().UsedFor(ArmTarget), // TODO: For now only used for ARM
		simplifyDefinitions(),
		injectJsonSerializationTests(idFactory).UsedFor(ArmTarget),

		markStorageVersion(),

		// Safety checks at the end:
		ensureDefinitionsDoNotUseAnyTypes(),

		deleteGeneratedCode(configuration.OutputPath),

		exportPackages(configuration.OutputPath).
			RequiresPrerequisiteStages("deleteGenerated"),
	}
}
//
//func crossplaneCorePipelineStages(idFactory astmodel.IdentifierFactory, configuration *config.Configuration) []PipelineStage {
//	return []PipelineStage{
//		// Import status info from Swagger:
//		augmentResourcesWithStatus(idFactory, configuration),
//
//		// Reduces oneOf/allOf types from schemas to object types:
//		convertAllOfAndOneOfToObjects(idFactory),
//
//		// Flatten out any nested resources created by allOf, etc. we want to do this before naming types or things
//		// get named with names like Resource_Spec_Spec_Spec:
//		flattenResources(), stripUnreferencedTypeDefinitions(),
//
//		// Name all anonymous object and enum types (required by controller-gen):
//		nameTypesForCRD(idFactory),
//
//		// Apply property type rewrites from the config file
//		// must come after nameTypesForCRD and convertAllOfAndOneOf so that objects are all expanded
//		applyPropertyRewrites(configuration),
//
//		// Figure out ARM resource owners:
//		determineResourceOwnership(),
//
//		// Strip out redundant type aliases:
//		removeTypeAliases(),
//
//		// De-pluralize resource types:
//		// improveResourcePluralization(),
//
//		stripUnreferencedTypeDefinitions(),
//
//		// Apply export filters before generating
//		// ARM types for resources etc:
//		applyExportFilters(configuration),
//		stripUnreferencedTypeDefinitions(),
//		replaceAnyTypeWithJSON(),
//
//		filterOutDefinitionsUsingAnyType(configuration.AnyTypePackages),
//
//		// createArmTypesAndCleanKubernetesTypes(idFactory),
//
//		addCrossplaneOwnerProperties(idFactory),
//		addCrossplaneForProvider(idFactory),
//		addCrossplaneAtProvider(idFactory),
//		addCrossplaneEmbeddedResourceSpec(idFactory),
//		addCrossplaneEmbeddedResourceStatus(idFactory),
//
//		// applyKubernetesResourceInterface(idFactory),
//		// createStorageTypes(),
//		simplifyDefinitions(),
//
//		// Safety checks at the end:
//		ensureDefinitionsDoNotUseAnyTypes(),
//		checkForMissingStatusInformation(),
//	}
//}

// Generate produces the Go code corresponding to the configured JSON schema in the given output folder
func (generator *CodeGenerator) Generate(ctx context.Context) error {
	klog.V(1).Infof("Generator version: %v", combinedVersion())

	defs := make(astmodel.Types)
	for i, stage := range generator.pipeline {
		klog.V(0).Infof("Pipeline stage %d/%d: %s", i+1, len(generator.pipeline), stage.description)
		// Defensive copy (in case the pipeline modifies its inputs) so that we can compare types in vs out
		defsOut, err := stage.action(ctx, defs.Copy())
		if err != nil {
			return errors.Wrapf(err, "failed during pipeline stage %d/%d: %s", i+1, len(generator.pipeline), stage.description)
		}

		defsAdded := defsOut.Except(defs)
		defsRemoved := defs.Except(defsOut)

		if len(defsAdded) > 0 && len(defsRemoved) > 0 {
			klog.V(1).Infof("Added %d, removed %d type definitions", len(defsAdded), len(defsRemoved))
		} else if len(defsAdded) > 0 {
			klog.V(1).Infof("Added %d type definitions", len(defsAdded))
		} else if len(defsRemoved) > 0 {
			klog.V(1).Infof("Removed %d type definitions", len(defsRemoved))
		}

		defs = defsOut
	}

	klog.Info("Finished")

	return nil
}

func (generator *CodeGenerator) verifyPipeline() error {
	var errs []error

	stagesSeen := make(map[string]struct{})
	for _, stage := range generator.pipeline {
		for _, prereq := range stage.prerequisites {
			if _, ok := stagesSeen[prereq]; !ok {
				errs = append(errs, errors.Errorf("prerequisite '%s' of stage '%s' not satisfied.", prereq, stage.id))
			}
		}

		stagesSeen[stage.id] = struct{}{}
	}

	return kerrors.NewAggregate(errs)
}
