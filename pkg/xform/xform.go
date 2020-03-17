/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

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
		Scheme *runtime.Scheme
	}

	idRef struct {
		ID string `json:"id,omitempty"`
	}
)

func NewARMConverter(client client.Client, scheme *runtime.Scheme) *ARMConverter {
	return &ARMConverter{
		Client: client,
		Scheme: scheme,
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

	if err := m.setResourceProperties(ctx, unObj, obj, res); err != nil {
		return nil, fmt.Errorf("unable to set Properties with: %w", err)
	}

	return res, nil
}

func (m *ARMConverter) FromResource(res *zips.Resource, obj azcorev1.MetaObject) error {
	unObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return fmt.Errorf("unable to convert to unstructured during ARM conversion: %w", err)
	}

	if err := m.setObjectStatus(res, unObj); err != nil {
		return fmt.Errorf("unable to set object status fields with: %w", err)
	}

	//if err := m.setObjectSpec(res, obj); err != nil {
	//	return fmt.Errorf("unable to set object spec fields with: %w", err)
	//}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(unObj, obj)
}

//func (m *ARMConverter) setObjectSpec(res *zips.Resource, obj azcorev1.MetaObject) error {
//	t := reflect.TypeOf(obj)
//
//	specField, found := t.FieldByName("Spec")
//	if !found {
//		return fmt.Errorf("unable to find Spec field on MetaObject %v", obj)
//	}
//
//	valueField := reflect.ValueOf(specField)
//	if valueField.Kind() != reflect.Struct {
//		return fmt.Errorf("expected Spec field value to be a struct %v", valueField)
//	}
//
//	bits, err := json.Marshal(res.Properties)
//	if err != nil {
//		return fmt.Errorf("unable to json.Marshal res.Properties with: %w", err)
//	}
//
//	if err := json.Unmarshal(bits, valueField.Pointer()); err != nil {
//		return fmt.Errorf("uanble to json.Unmarshal res.Properites into Spec field with: %w", err)
//	}
//
//	return nil
//}

func (m *ARMConverter) setObjectStatus(res *zips.Resource, unObj map[string]interface{}) error {
	if err := unstructured.SetNestedField(unObj, res.ID, "status", "id"); err != nil {
		return fmt.Errorf("unable to set status.id with: %w", err)
	}

	if err := unstructured.SetNestedField(unObj, res.DeploymentID, "status", "deploymentId"); err != nil {
		return fmt.Errorf("unable to set status.deploymentId with: %w", err)
	}

	if err := unstructured.SetNestedField(unObj, string(res.ProvisioningState), "status", "provisioningState"); err != nil {
		return fmt.Errorf("unable to set status.provisioningState with: %w", err)
	}

	return nil
}

func (m *ARMConverter) checkAllOwnersSucceeded(ctx context.Context, obj azcorev1.MetaObject) (bool, error) {
	// has owners, so check to see if those owners are ready
	for _, ref := range obj.GetOwnerReferences() {
		owner, err := m.Scheme.New(schema.GroupVersionKind{
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
		if err != nil {
			return false, fmt.Errorf("unable to convert to unstructured with: %w", err)
		}

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

func (m *ARMConverter) setResourceProperties(ctx context.Context, unObj map[string]interface{}, obj azcorev1.MetaObject, res *zips.Resource) error {
	refs, err := GetTypeReferenceData(obj)
	if err != nil {
		return fmt.Errorf("unable to gather type reference tags with: %w", err)
	}

	for _, ref := range refs {
		var err error
		if ref.IsSlice {
			err = m.replaceSliceReferenceWithIDs(ctx, unObj, obj, ref)
			if err != nil {
				err = fmt.Errorf("failed to replace slice reference with IDs with: %w", err)
			}
		} else {
			err = m.replaceReferenceWithID(ctx, unObj, obj, ref)
			if err != nil {
				err = fmt.Errorf("failed to replace reference with ID with: %w", err)
			}
		}

		if err != nil {
			return err
		}
	}

	unProps, found, err := unstructured.NestedMap(unObj, "spec", "properties")
	if err != nil {
		return fmt.Errorf("unable to fetch unstructured obj.spec.properties with: %w", err)
	}

	if !found {
		return nil
	}

	var raw json.RawMessage
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unProps, &raw); err != nil {
		return fmt.Errorf("unable to marshal unstructured properties to json with: %w", err)
	}

	res.Properties = raw
	return nil
}

func (m *ARMConverter) replaceReferenceWithID(ctx context.Context, unObj map[string]interface{}, obj azcorev1.MetaObject, ref TypeReferenceLocation) error {
	unRef, found, err := unstructured.NestedMap(unObj, ref.JSONFields()...)
	if err != nil {
		return fmt.Errorf("unable to find path %v with: %w", ref.JSONFields(), err)
	}

	if !found {
		// ref was not found, no need to replace it
		return nil
	}

	// remove the KnownTypeReference
	unstructured.RemoveNestedField(unObj, ref.JSONFields()...)

	var knownTypeRef azcorev1.KnownTypeReference
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unRef, &knownTypeRef); err != nil {
		return fmt.Errorf("unable to build KnownTypeReference from unstructured with: %w", err)
	}

	if knownTypeRef.Name == "" {
		// name of the reference is not set, so we will ignore it
		return nil
	}

	if knownTypeRef.Namespace == "" {
		// default to the current object's namespace if not specified on the reference
		knownTypeRef.Namespace = obj.GetNamespace()
	}

	nn := client.ObjectKey{
		Name:      knownTypeRef.Name,
		Namespace: knownTypeRef.Namespace,
	}

	gvk := schema.GroupVersionKind{
		Group:   obj.GetObjectKind().GroupVersionKind().Group,
		Version: "v1",
		Kind:    ref.Kind,
	}

	refObj, err := m.Scheme.New(gvk)
	if err != nil {
		return fmt.Errorf("unable to find gvk for ref %v with: %w", ref, err)
	}

	if err := m.Client.Get(ctx, nn, refObj); err != nil {
		return fmt.Errorf("unable to fetch object %v with: %w", nn, err)
	}

	unRefObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(refObj)
	if err != nil {
		return fmt.Errorf("unable to convert refObj to unstructured with: %w", err)
	}

	id, found, err := unstructured.NestedString(unRefObj, "status", "id")
	if err != nil {
		return fmt.Errorf("unable to find unRefObj.status.id with: %v", err)
	}

	if found {
		idMap := map[string]interface{}{
			"id": id,
		}
		if err := unstructured.SetNestedMap(unObj, idMap, ref.TemplateFields()...); err != nil {
			return fmt.Errorf("unable to set ID map for reference %v with: %w", ref, err)
		}
	}

	return nil
}

func (m *ARMConverter) replaceSliceReferenceWithIDs(ctx context.Context, unObj map[string]interface{}, obj azcorev1.MetaObject, ref TypeReferenceLocation) error {
	unRef, found, err := unstructured.NestedSlice(unObj, ref.JSONFields()...)
	if err != nil {
		return fmt.Errorf("unable to find path %v with: %w", ref.JSONFields(), err)
	}

	if !found {
		// ref was not found, no need to replace it
		return nil
	}

	// remove the KnownTypeReference
	unstructured.RemoveNestedField(unObj, ref.JSONFields()...)

	var knownTypeRefsMap map[string][]azcorev1.KnownTypeReference
	unRefMap := map[string]interface{}{
		"ktrs": unRef,
	}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unRefMap, &knownTypeRefsMap); err != nil {
		return fmt.Errorf("unable to build KnownTypeReference from unstructured with: %w", err)
	}

	var ids []interface{}
	knownTypeRefs := knownTypeRefsMap["ktrs"]
	for _, ktr := range knownTypeRefs {
		if ktr.Name == "" {
			// name of the reference is not set, so we will ignore it
			continue
		}

		if ktr.Namespace == "" {
			// default to the current object's namespace if not specified on the reference
			ktr.Namespace = obj.GetNamespace()
		}

		nn := client.ObjectKey{
			Name:      ktr.Name,
			Namespace: ktr.Namespace,
		}

		gvk := schema.GroupVersionKind{
			Group:   obj.GetObjectKind().GroupVersionKind().Group,
			Version: "v1",
			Kind:    ref.Kind,
		}

		refObj, err := m.Scheme.New(gvk)
		if err != nil {
			return fmt.Errorf("unable to find gvk for ref %v with: %w", ref, err)
		}

		if err := m.Client.Get(ctx, nn, refObj); err != nil {
			return fmt.Errorf("unable to fetch object %v with: %w", nn, err)
		}

		unRefObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(refObj)
		if err != nil {
			return fmt.Errorf("unable to convert refObj to unstructured with: %w", err)
		}

		id, found, err := unstructured.NestedString(unRefObj, "status", "id")
		if err != nil {
			return fmt.Errorf("unable to find unRefObj.status.id with: %v", err)
		}

		if found {
			unId, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&idRef{
				ID: id,
			})

			if err != nil {
				return fmt.Errorf("unable to convert idRef to unstructured with: %w", err)
			}
			ids = append(ids, unId)
		}
	}

	if err := unstructured.SetNestedSlice(unObj, ids, ref.TemplateFields()...); err != nil {
		return fmt.Errorf("unable to set nested slice of IDs with: %w", err)
	}

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

	state, ok, err := unstructured.NestedString(status, "provisioningState")
	if err != nil {
		return fmt.Errorf("unable to extract provisioningState from status with: %w", err)
	}

	if ok {
		res.ProvisioningState = zips.ProvisioningState(state)
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
