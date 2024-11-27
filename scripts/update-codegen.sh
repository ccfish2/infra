#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/.."
GOPATH=/Users/jiminhu/go
CODEGEN_PKG=$GOPATH/src/github.com/code-generator

source "${CODEGEN_PKG}/kube_codegen.sh"

TMPDIR=${1}
PLURAL_EXCEPTIONS="DolphinEndpoints:DolphinEndpoints,DolphinEnvoyConfig:DolphinEnvoyConfigs,DolphinEndpointSlice:DolphinEndpointSlices"

kube::codegen::gen_client \
    "./pkg/k8s/apis" \
    --with-watch \
    --output-dir "${TMPDIR}/github.com/ccfish2/infra/pkg/k8s/client" \
    --output-pkg "github.com/ccfish2/infra/pkg/k8s/client" \
    --plural-exceptions ${PLURAL_EXCEPTIONS} \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt"

cp -r "${TMPDIR}/github.com/ccfish2/infra/." ./

kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt" \
    "$PWD/pkg/k8s/apis"