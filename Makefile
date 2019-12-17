SHELL := /bin/bash

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Directories.
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

# Binaries.
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/controller-gen
KUBECTL=$(TOOLS_BIN_DIR)/kubectl
KUBE_APISERVER=$(TOOLS_BIN_DIR)/kube-apiserver
ETCD=$(TOOLS_BIN_DIR)/etcd

all: manager

## --------------------------------------
## Testing
## --------------------------------------

test: export TEST_ASSET_KUBECTL = $(ROOT_DIR)/$(KUBECTL)
test: export TEST_ASSET_KUBE_APISERVER = $(ROOT_DIR)/$(KUBE_APISERVER)
test: export TEST_ASSET_ETCD = $(ROOT_DIR)/$(ETCD)

# Run tests
test: $(KUBECTL) $(KUBE_APISERVER) $(ETCD) generate fmt vet manifests
	go test ./... -coverprofile cover.out

$(KUBECTL) $(KUBE_APISERVER) $(ETCD): ## install test asset kubectl, kube-apiserver, etcd
	. ./scripts/fetch_ext_bins.sh && fetch_tools

$(CONTROLLER_GEN): $(TOOLS_DIR)/go.mod # Build controller-gen from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o bin/controller-gen sigs.k8s.io/controller-tools/cmd/controller-gen

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}
