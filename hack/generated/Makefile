SHELL             = /bin/bash
PACKAGE           = github.com/Azure/k8s-infra/hack/generated
APP               = k8sinfra-controller

# timestamp         := $(shell /bin/date "+%Y%m%d-%H%M%S")
CONFIG_REGISTRY = kind-registry:5000/fake/k8s-infra-controller:latest
# IMG               ?= k8s-infra-generated-contoller:$(timestamp)
IMG               ?= k8s-infra-generated-contoller:latest
KIND_CLUSTER_NAME = k8sinfra-generated

include ../../tools.mk

CRD_OPTIONS     ?= "crd:crdVersions=v1,allowDangerousTypes=true"
GO_DIRS         := $(shell $(GO) list -f '{{.Dir}}' ./...)
# We exclude the apis folder because it's really large and test discovery takes a good amount of time (>10s)
GO_DIRS_TO_TEST := $(shell $(GO) list -f '{{.Dir}}' ./... | grep -v /apis/)
CONTROLLER_DEBUG_LOGLEVEL := 4

V = 0
Q = $(if $(filter 1,$V),,@)

.PHONY: all
all: generate header-check fmt build test test-int

# There is a ci specific target because we want the CI pass to fail if
# the code has not been go fmt-ed, whereas locally we want "make all"
# to just format the code for you
.PHONY: ci
ci: generate build test-cover test-int

.PHONY: lint
lint: $(GOLANGCI_LINT) ; $(info $(M) running golangci configured linters…) ## Lint codebase
	$(Q) $(GOLANGCI_LINT) run -v

.PHONY: lint-full
lint-full: $(GOLANGCI_LINT) ; $(info $(M) running golangci configured linters…) ## Run slower linters to detect possible issues
	$(Q) $(GOLANGCI_LINT) run -v --fast=false

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	$(Q) $(GOFMT) -l -s -w $(GO_DIRS)

.PHONY: tidy
tidy: ; $(info $(M) running tidy…) @ ## Run tidy
	$Q $(GO) mod tidy

.PHONY: test
test: ; $(info $(M) running go test…)
	$(Q) $(GO) test $(GO_DIRS_TO_TEST) -tags=noexit

# Initially this target uses kind as there is no need to end to end test on Azure.
# Eventually when we start supporting features like MSI that only work in Azure,
# it's likely we'll need to make the integration tests (or at least some of them)
# run in an AKS cluster.
.PHONY: test-int-no-cleanup
test-int-no-cleanup: $(ROOT_DIR)/.env kind-create deploy ; $(info $(M) running controller integration test…)
	# MUST be executed as single command, or env vars will not propagate to test execution
	$(Q) . $(ROOT_DIR)/.env && $(GO) test $(GO_DIRS_TO_TEST) -tags integration # TODO: better way to just run integration tests?

.PHONY: test-int
test-int: test-int-no-cleanup kind-delete

.PHONY: test-cover
test-cover: $(GCOV2LCOV) ; $(info $(M) running go test…)
	# NOTE: if you update the 'test-cover' target, also update ./github/workflows/test.yml
	#       for the Windows part of the "test-generator" job.
	$(Q) $(GO) test -tags=noexit -race -covermode atomic -coverprofile=cover.out -coverpkg=./... $(GO_DIRS_TO_TEST)
	$(Q) $(GCOV2LCOV) -infile cover.out -outfile coverage.lcov

.PHONY: generate
generate: $(CONTROLLER_GEN) $(CONVERSION_GEN) ## Generate code
	# Force regeneration of all of the conversions
	@echo "Deleting old deepcopy files"
	$(Q) find "./apis" -type f -name "zz_generated.*" -delete

	@echo "Executing controller-gen to generate deepcopy functions"
	$(Q) $(CONTROLLER_GEN) object:headerFile=../boilerplate.go.txt paths="./..."

#	@echo "Executing conversion-gen"
#	$(Q) $(CONVERSION_GEN) \
#		--input-dirs="./apis/..." \
#		--output-file-base=zz_generated.conversion \
#		--output-base=./ \
#		--go-header-file=../boilerplate.go.txt

	# Force regeneration of all of the CRDs
	@echo "Deleting old CRD YAMLs"
	$(Q) if [ -d "./config/crd/bases" ]; then find "./config/crd/bases" -type f -name "*" -delete; fi

	@echo "Executing controller-gen to generate CRD and RBAC YAMLs"
	$(Q) $(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: build
build: tidy lint ; $(info $(M) building ./bin/$(APP))
	$(Q) $(GO) build -o ./bin/$(APP)

## --------------------------------------
## Development
## --------------------------------------

.PHONY: kind-delete
kind-delete: $(KIND) ## Destroys the "k8sinfra" kind cluster.
	$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME) || true

.PHONY: kind-create
kind-create:
	export KIND_CLUSTER_NAME=$(KIND_CLUSTER_NAME) && $(SCRIPTS_DIR)/kind-with-registry.sh

# TODO: We may want this later
#.PHONY: apply-certs-and-secrets
#apply-certs-and-secrets: $(KUBECTL) ## Apply certificate manager and manager secrets into cluster
#	$(SCRIPTS_DIR)/apply_cert_and_secrets.sh

.PHONY: run
run: export ENVIRONMENT = development
run: kind-create install ## Run a development cluster using kind
	$(Q) $(GO) run ./main.go -v $(CONTROLLER_DEBUG_LOGLEVEL)

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
	# TODO: Consider patching in CONTROLLER_DEBUG_LOGLEVEL?
	$(KUSTOMIZE) build config/default | sed "s_${CONFIG_REGISTRY}_${REGISTRY}/${IMG}_" | $(KUBECTL) apply -f -
	# TODO: Is there a better way to do the below?
	$(SCRIPTS_DIR)/deploy_testing_secret.sh

.PHONY: docker-build
docker-build: ## Build the docker image
	docker build . -t $(REGISTRY)/${IMG}

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(REGISTRY)/${IMG}
