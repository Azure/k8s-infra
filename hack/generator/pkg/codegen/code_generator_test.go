package codegen

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"

	. "github.com/onsi/gomega"
)

// Remove all stages from the pipeline with the given ids
// Will panic if you specify an unknown id
func (generator *CodeGenerator) RemoveStages(stageIds ...string) {
	stagesToRemove := make(map[string]bool)
	for _, s := range stageIds {
		stagesToRemove[s] = false
	}

	var pipeline []PipelineStage

	for _, stage := range generator.pipeline {
		if _, ok := stagesToRemove[stage.id]; ok {
			stagesToRemove[stage.id] = true
			continue
		}

		pipeline = append(pipeline, stage)
	}

	for stage, removed := range stagesToRemove {
		if !removed {
			panic(fmt.Sprintf("Expected to remove stage %s from pipeline, but it wasn't found.", stage))
		}
	}

	generator.pipeline = pipeline
}

// Replace all uses of an existing stage with another one
// Will panic if the existing stage is not found
func (generator *CodeGenerator) ReplaceStage(existingStage string, stage PipelineStage) {
	replaced := false
	for i, s := range generator.pipeline {
		if s.HasId(existingStage) {
			generator.pipeline[i] = stage
			replaced = true
		}
	}

	if !replaced {
		panic(fmt.Sprintf("Expected to replace stage %s but it wasn't found", existingStage))
	}
}

func (generator *CodeGenerator) HasStage(id string) bool {
	for _, s := range generator.pipeline {
		if s.HasId(id) {
			return true
		}
	}

	return false
}

/*
 * Shared test data
 */

var (
	fooStage = MakeFakePipelineStage("foo")
	barStage = MakeFakePipelineStage("bar")
	bazStage = MakeFakePipelineStage("baz")
	zooStage = MakeFakePipelineStage("zoo")
)

/*
 * RemoveStagesTests
 */

func TestRemoveStages_RemovesSpecifiedStages(t *testing.T) {
	g := NewGomegaWithT(t)

	gen := &CodeGenerator{
		pipeline: []PipelineStage{
			fooStage,
			barStage,
			bazStage,
		},
	}

	gen.RemoveStages("foo", "baz")
	g.Expect(gen.pipeline).To(HaveLen(1))
	g.Expect(gen.pipeline[0].HasId("bar")).To(BeTrue())
}

func TestRemoveStages_PanicsForUnknownStage(t *testing.T) {
	g := NewGomegaWithT(t)

	gen := &CodeGenerator{
		pipeline: []PipelineStage{
			fooStage,
			barStage,
			bazStage,
		},
	}

	g.Expect(func() {
		gen.RemoveStages("bang")
	},
	).To(Panic())

	gen.RemoveStages("foo", "baz")
}

func MakeFakePipelineStage(id string) PipelineStage {
	return MakePipelineStage(
		id, "Stage "+id, func(ctx context.Context, types astmodel.Types) (astmodel.Types, error) {
			return types, nil
		})
}

/*
 * ReplaceStageTests
 */

func TestReplaceStage_ReplacesSpecifiedStage(t *testing.T) {
	g := NewGomegaWithT(t)

	gen := &CodeGenerator{
		pipeline: []PipelineStage{
			fooStage,
			barStage,
			bazStage,
		},
	}

	gen.ReplaceStage("foo", zooStage)
	g.Expect(gen.pipeline).To(HaveLen(3))
	g.Expect(gen.HasStage("foo")).To(BeFalse())
	g.Expect(gen.HasStage("zoo")).To(BeTrue())
}

func TestReplaceStage_PanicsForUnknownStage(t *testing.T) {
	g := NewGomegaWithT(t)

	gen := &CodeGenerator{
		pipeline: []PipelineStage{
			fooStage,
			barStage,
			bazStage,
		},
	}

	g.Expect(func() {
		gen.ReplaceStage("bang", zooStage)
	},
	).To(Panic())
}
