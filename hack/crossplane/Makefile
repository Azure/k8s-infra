SHELL             = /bin/bash
PACKAGE           = github.com/Azure/k8s-infra/hack/generated
APP               = crossplane-test

# timestamp         := $(shell /bin/date "+%Y%m%d-%H%M%S")
# CONFIG_REGISTRY = kind-registry:5000/fake/k8s-infra-controller:latest
# IMG               ?= k8s-infra-generated-contoller:$(timestamp)
# KIND_CLUSTER_NAME = k8sinfra-generated

include ../../tools.mk

CRD_OPTIONS     ?= "crd:crdVersions=v1,allowDangerousTypes=true"
GO_DIRS         := $(shell $(GO) list -f '{{.Dir}}' ./...)
# We exclude the apis folder because it's really large and test discovery takes a good amount of time (>10s)
GO_DIRS_TO_TEST := $(shell $(GO) list -f '{{.Dir}}' ./... | grep -v /apis/)

V = 0
Q = $(if $(filter 1,$V),,@)

.PHONY: all
all: generate header-check fmt build test

# There is a ci specific target because we want the CI pass to fail if
# the code has not been go fmt-ed, whereas locally we want "make all"
# to just format the code for you
.PHONY: ci
ci: generate build test-cover

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

.PHONY: test-cover
test-cover: $(GCOV2LCOV) ; $(info $(M) running go test…)
	# NOTE: if you update the 'test-cover' target, also update ./github/workflows/test.yml
	#       for the Windows part of the "test-generator" job.
	$(Q) $(GO) test -tags=noexit -race -covermode atomic -coverprofile=cover.out -coverpkg=./... $(GO_DIRS_TO_TEST)
	$(Q) $(GCOV2LCOV) -infile cover.out -outfile coverage.lcov

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

	# Force regeneration of all of the CRDs
	@echo "Deleting old CRDs"
	$(Q) if [ -d "./config/crd/bases" ]; then find "./config/crd/bases" -type f -name "*" -delete; fi

	@echo "Executing controller-gen to generate CRDs"
	$(Q) $(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./apis/..." output:crd:artifacts:config=config/crd/bases

.PHONY: build
build: tidy lint ; $(info $(M) building ./bin/$(APP))
	$(Q) $(GO) build "./apis/..." -o ./bin/$(APP)

