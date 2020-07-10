SHELL             = /bin/bash
PACKAGE           = github.com/Azure/k8s-infra/hack/generated
APP               = k8sinfra-controller

timestamp         := $(shell /bin/date "+%Y%m%d-%H%M%S")
# CONFIG_REGISTRY = kind-registry:5000/fake/k8s-infra-controller:latest
IMG               ?= k8s-infra-generated-contoller:$(timestamp)
KIND_CLUSTER_NAME = k8sinfra-generated

include ../../tools.mk

CRD_OPTIONS     ?= "crd:crdVersions=v1,allowDangerousTypes=true"

V = 0
Q = $(if $(filter 1,$V),,@)

.PHONY: all
all: generate build

.PHONY: generate
generate: $(CONTROLLER_GEN) $(CONVERSION_GEN) ## Generate code
	# Force regeneration of all of the conversions
	@echo "Deleting old controller gen files"
	$(Q) find "./apis" -type f -name "zz_generated.*" -delete

	@echo "Executing controller-gen"
	$(Q) $(CONTROLLER_GEN) object:headerFile=../boilerplate.go.txt paths="./..."

#	@echo "Executing conversion-gen"
#	$(Q) $(CONVERSION_GEN) \
#		--input-dirs="./apis/..." \
#		--output-file-base=zz_generated.conversion \
#		--output-base=./ \
#		--go-header-file=../boilerplate.go.txt

	@echo "Executing controller-gen to generate CRDs"
	$(Q) $(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./apis/..." output:crd:artifacts:config=config/crd/bases

# TODO: Right now this is just proving that the packages can be built - we don't have any executable consuming them yet so no binary is produced
.PHONY: build
build:
	$(Q) $(GO) build -o ./bin/$(APP)

## --------------------------------------
## Development
## --------------------------------------

.PHONY: kind-delete
kind-delete: $(KIND) ## Destroys the "k8sinfra" kind cluster.
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME) || true

.PHONY: kind-create
kind-create:
	$(Q) $(KIND) get clusters | grep -E $(KIND_CLUSTER_NAME) > /dev/null;\
	EXISTS=$$?;\
	if [ $$EXISTS -eq 0 ]; then \
		echo "$(KIND_CLUSTER_NAME) already exists"; \
	else \
		$(KIND) create cluster --name=$(KIND_CLUSTER_NAME); \
	fi; \

	# TODO: Need to use this script when we actually start installing stuff from a registry
	# $(SCRIPTS_DIR)/kind-with-registry.sh

#.PHONY: apply-certs-and-secrets
#apply-certs-and-secrets: $(KUBECTL) ## Apply certificate manager and manager secrets into cluster
#	$(SCRIPTS_DIR)/apply_cert_and_secrets.sh

.PHONY: run
# run: export KUBECONFIG = $(KIND_KUBECONFIG)
run: export ENVIRONMENT = development
run: kind-create install ## Run a development cluster using kind
	$(GO) run ./main.go

## --------------------------------------
## Deploy
## --------------------------------------

.PHONY: install
install: generate $(KUBECTL) $(KUSTOMIZE) ## Install CRDs into a cluster
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: generate $(KUBECTL) $(KUSTOMIZE) ## Uninstall CRDs from a cluster
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete -f -

.PHONY: deploy
deploy: generate $(KUBECTL) $(KUSTOMIZE) docker-build docker-push ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	$(KUSTOMIZE) build config/default | sed "s_${CONFIG_REGISTRY}_${REGISTRY}/${IMG}_" | $(KUBECTL) apply -f -

.PHONY: docker-build
docker-build: ## Build the docker image
	docker build . -t $(REGISTRY)/${IMG}

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(REGISTRY)/${IMG}