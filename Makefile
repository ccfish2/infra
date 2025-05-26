ifeq ($(ROOT_DIR),)
	ROOT_DIR ?= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
	RELATIVE_DIR ?= $(shell echo $(realpath .) | sed "s;$(ROOT_DIR)[/]*;;")
endif
ifeq ($(V),0)
	QUIET=@
	ECHO_CC=echo "  CC     $(RELATIVE_DIR)/$@"
	ECHO_CHECK=echo "  CHECK  $(RELATIVE_DIR)"
	ECHO_CLEAN=echo "  CLEAN  $(RELATIVE_DIR)"
	ECHO_DOCKER=echo "  DOCKER $(RELATIVE_DIR) $@"
	ECHO_GEN=echo "  GEN    $(RELATIVE_DIR)/"
	ECHO_GINKGO=echo "  GINKGO $(RELATIVE_DIR)"
	ECHO_GO=echo "  GO     $(RELATIVE_DIR)/$@"
	ECHO_TEST=echo "  TEST "
	SUBMAKEOPTS="-s"
else
	# The whitespace at below EOLs is required for verbose case!
	ECHO_CC=: 
	ECHO_CHECK=: 
	ECHO_CLEAN=: 
	ECHO_DOCKER=: 
	ECHO_GEN=: 
	ECHO_GINKGO=: 
	ECHO_GO=: 
	ECHO_TEST=: 
	SUBMAKEOPTS=
endif

export GO ?= go
NATIVE_ARCH = $(shell GOARCH= $(GO) env GOARCH)
export GOARCH ?= $(NATIVE_ARCH)

##@ API targets
CRD_OPTIONS ?= "crd:crdVersions=v1"
CRD_IMPORT_PATH := github.com/ccfish2/infra/pkg/k8s/apis/dolphin.io/v1
CRD_FS_PATH := $(PWD)/pkg/k8s/apis/dolphin.io/v1
CRDS_DOLPHIN_OUTDIR := ./pkg/k8s/apis/dolphin.io/client/crd/bases
CRDS_DOLPHIN_V1 := dolphinendpoints \
				   dolphinendpointslices \
                   dolphinidentities \
                   dolphinenvoyconfigs \
				   dolphinnodes

manifests: ## Generate K8s manifests e.g. CRD, RBAC etc.
	$(eval TMPDIR := $(shell mktemp -d -t dolphin.tmpXXXXXXXX))
	$(QUIET)$(GO) run sigs.k8s.io/controller-tools/cmd/controller-gen $(CRD_OPTIONS) paths=$(CRD_IMPORT_PATH) output:crd:artifacts:config="$(TMPDIR)"
	$(QUIET)$(GO) run ./tools/crdcheck "$(TMPDIR)"

	# Clean up old CRD state and start with a blank state.
	rm -rf $(CRDS_DOLPHIN_OUTDIR)
	mkdir -p $(CRDS_DOLPHIN_OUTDIR)

	for file in $(CRDS_DOLPHIN_V1); do \
		mv ${TMPDIR}/dolphin.io_$$file.yaml $(CRDS_DOLPHIN_OUTDIR)/$$file.yaml; \
	done
	
	rm -rf $(TMPDIR)

generate-k8s-api: ## Generate ccfish2 infra k8s API client, deepcopy and deepequal Go sources.

	$(eval TMPDIR := $(shell mktemp -d -t infra.tmpXXXXXXXX))

	$(eval DEEPEQUAL_PACKAGES := $(shell grep "\+deepequal-gen" -l -r --include \*.go --exclude-dir 'vendor' . | xargs dirname {} | sort | uniq | grep -x -v '.' | sed 's|\.\/|github.com/ccfish2/infra\/|g'))
	# $(QUIET) $(call generate_deepequal,${DEEPEQUAL_PACKAGES})

	$(eval DEEPCOPY_PACKAGES := $(shell grep "\+k8s:deepcopy-gen" -l -r --include \*.go --exclude-dir 'vendor' . | xargs dirname {} | sort | uniq | grep -x -v '.' | sed 's|\.\/|github.com/ccfish2/infra\/|g'))
	$(QUIET) $(call generate_deepcopy,${DEEPCOPY_PACKAGES},"$(TMPDIR)")

	$(QUIET) $(call generate_k8s_api,client$(comma)lister$(comma)informer,github.com/ccfish2/infra/pkg/k8s/client,github.com/ccfish2/infra/pkg/k8s/apis,$(GEN_CRD_GROUPS),"$(TMPDIR)")

	$(QUIET) cp -r "$(TMPDIR)/github.com/ccfish2/infra/." ./
	$(QUIET) rm -rf "$(TMPDIR)"

define generate_k8s_api
	sudo scripts/update-codegen.sh /Users/jiminhu/go/src
endef

define generate_deepequal
	$(GO) run github.com/cilium/deepequal-gen \
	--input-dirs $(subst $(space),$(comma),$(1)) \
	--go-header-file "$(PWD)/hack/custom-boilerplate.go.txt" \
	--output-file-base zz_generated.deepequal \
	--output-base $(2)
endef

define generate_deepcopy
	$(GO) run k8s.io/code-generator/cmd/deepcopy-gen \
	--input-dirs $(subst $(space),$(comma),$(1)) \
	--go-header-file "$(PWD)/hack/custom-boilerplate.go.txt" \
	--output-file-base zz_generated.deepcopy \
	--output-base $(2)
endef

gofmt: ## Run gofmt on Go source files in the repository.
	$(QUIET)$(GO) fmt ./...

govet: ## Run govet on Go source files in the repository.
	@$(ECHO_CHECK) vetting all packages...
	$(QUIET) $(GO_VET) ./...
