package xform

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	azcorev1 "github.com/Azure/k8s-infra/apis/core/v1"
	"github.com/Azure/k8s-infra/pkg/zips"
)

type (
	ARMConverter struct {
		Client client.Client
		Schema *runtime.Scheme
	}
)

func NewARMConverter(client client.Client, schema *runtime.Scheme) *ARMConverter {
	return &ARMConverter{
		Client: client,
		Schema: schema,
	}
}

func (m *ARMConverter) ToResource(ctx context.Context, obj azcorev1.MetaObject) (*zips.Resource, error) {
	unObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to unstructured during ARM conversion: %w", err)
	}

	res := new(zips.Resource)
	res.SetAnnotations(obj.GetAnnotations())
	res.Name = obj.GetName()
	res.Type = obj.ResourceType()
	if grouped, ok := obj.(azcorev1.Grouped); ok {
		groupRef := grouped.GetResourceGroupObjectRef()
		res.ResourceGroup = groupRef.Name
	}

	ok, err := m.checkAllOwnersSucceeded(ctx, obj)
	if err != nil {
		return res, err
	}

	if !ok {
		return res, fmt.Errorf("an owner reference is not in a Succeeded provisioning state")
	}

	if err := setTopLevelResourceFields(unObj, res); err != nil {
		return res, err
	}

	spec, ok, err := unstructured.NestedMap(unObj, "spec")
	if err != nil {
		return nil, fmt.Errorf("unable to extract next spec map with: %w", err)
	}

	if !ok {
		return nil, fmt.Errorf("spec not found in object")
	}

	if err := setProperties(ctx, spec, obj); err != nil {
		return nil, fmt.Errorf("unable to set Properties with: %w", err)
	}

	return res, nil
}

func (m *ARMConverter) checkAllOwnersSucceeded(ctx context.Context, obj azcorev1.MetaObject) (bool, error) {
	// has owners, so check to see if those owners are ready
	for _, ref := range obj.GetOwnerReferences() {
		owner, err := m.Schema.New(schema.GroupVersionKind{
			Group:   obj.GetObjectKind().GroupVersionKind().Group,
			Version: "v1",
			Kind:    ref.Kind,
		})
		if err != nil {
			return false, fmt.Errorf("unable to find gvk for ref %v with: %w", ref, err)
		}

		refKey := client.ObjectKey{
			Name:      ref.Name,
			Namespace: obj.GetNamespace(),
		}

		if err := m.Client.Get(ctx, refKey, owner); err != nil {
			return false, fmt.Errorf("unable to fetch owner %v with: %w", ref, err)
		}

		unOwn, err := runtime.DefaultUnstructuredConverter.ToUnstructured(owner)
		state, ok, err := unstructured.NestedString(unOwn, "status", "provisioningState")
		if err != nil {
			return false, fmt.Errorf("error fetching unstructured provisioningState with: %w", err)
		}

		if ok {
			if zips.SucceededProvisioningState != zips.ProvisioningState(state) {
				return false, nil
			}
		}
	}

	return true, nil
}

func setProperties(ctx context.Context, spec map[string]interface{}, obj azcorev1.MetaObject) error {
	_, err := GetTypeReferenceData(obj)
	if err != nil {
		return fmt.Errorf("unable to gather type reference tags with: %w", err)
	}

	// use the refs to gather IDs for references

	// replace the azcorev1.KnownReferences with the ID references in the unstructured spec for template generation

	return nil
}

func setTopLevelResourceFields(unObj map[string]interface{}, res *zips.Resource) error {
	spec, ok, err := unstructured.NestedMap(unObj, "spec")
	if err != nil {
		return fmt.Errorf("unable to extract spec map with: %w", err)
	}

	if !ok {
		return fmt.Errorf("spec not found in object")
	}

	tags, ok, err := unstructured.NestedStringMap(spec, "tags")
	if err != nil {
		return fmt.Errorf("unable to extract tags from spec with: %w", err)
	}

	if ok {
		res.Tags = tags
	}

	apiVersion, err := getRequiredStringValue(spec, "apiVersion")
	if err != nil {
		return err
	}
	res.APIVersion = apiVersion

	managedBy, err := getStringValue(spec, "managedBy")
	if err != nil {
		return err
	}
	res.ManagedBy = managedBy

	location, err := getStringValue(spec, "location")
	if err != nil {
		return err
	}
	res.Location = location

	status, ok, err := unstructured.NestedMap(unObj, "status")
	if err != nil {
		return fmt.Errorf("unable to extract status map with: %w", err)
	}

	if !ok {
		return fmt.Errorf("status not found in object")
	}

	bits, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("unable to marshal status to json with: %w", err)
	}

	err = json.Unmarshal(bits, res)
	if err != nil {
		return fmt.Errorf("unable to unmarshal status to resource with: %w", err)
	}

	return nil
}

func getRequiredStringValue(unObj map[string]interface{}, fieldName string) (string, error) {
	str, ok, err := unstructured.NestedString(unObj, fieldName)
	if err != nil {
		return "", fmt.Errorf("unable to extract %s with: %w", fieldName, err)
	}

	if !ok {
		return "", fmt.Errorf("%s not found in spec", fieldName)
	}

	return str, nil
}

func getStringValue(unObj map[string]interface{}, fieldName string) (string, error) {
	str, ok, err := unstructured.NestedString(unObj, fieldName)
	if err != nil {
		return "", fmt.Errorf("unable to extract %s with: %w", fieldName, err)
	}

	if !ok {
		return "", nil
	}

	return str, nil
}
