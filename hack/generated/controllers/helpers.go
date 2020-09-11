/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Azure/k8s-infra/hack/generated/pkg/genruntime"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"time"
)

// SpecSignature calculates the hash of a spec. This can be used to compare specs and determine
// if there has been a change
func SpecSignature(metaObject genruntime.MetaObject) (string, error) {
	// Convert the resource to unstructured for easier comparison later.
	unObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(metaObject)
	if err != nil {
		return "", err
	}

	spec, ok, err := unstructured.NestedMap(unObj, "spec")
	if err != nil {
		return "", err
	}

	if !ok {
		return "", errors.New("unable to find spec within unstructured MetaObject")
	}

	bits, err := json.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("unable to marshal spec of unstructured MetaObject with: %w", err)
	}

	hash := sha256.Sum256(bits)
	return hex.EncodeToString(hash[:]), nil
}

// CreateDeploymentName generates a unique deployment name
func CreateDeploymentName() (string, error) {
	// no status yet, so start provisioning
	deploymentUUID, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	deploymentName := fmt.Sprintf("%s_%d_%s", "k8s", time.Now().Unix(), deploymentUUID.String())
	return deploymentName, nil
}
