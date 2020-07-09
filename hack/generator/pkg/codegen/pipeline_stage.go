package codegen

import "github.com/Azure/k8s-infra/hack/generator/pkg/astmodel"

// PipelineStage represents a composable stage of processing that can transform or process the set
// of generated types
type PipelineStage struct {
	Name string
	Action func ([]astmodel.TypeDefiner) ([]astmodel.TypeDefiner, error)
}
