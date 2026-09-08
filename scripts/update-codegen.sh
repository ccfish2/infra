#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/.."
GOPATH=${GOPATH:-$(go env GOPATH)}
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${GOPATH}/pkg/mod/k8s.io && ls -d code-generator* | head -1)}

source "${GOPATH}/pkg/mod/${CODEGEN_PKG}/kube_codegen.sh"

TMPDIR=${TMPDIR:-/tmp/codegen}
mkdir -p "${TMPDIR}"

PLURAL_EXCEPTIONS="DolphinEndpoints:DolphinEndpoints,DolphinEnvoyConfig:DolphinEnvoyConfigs,DolphinEndpointSlice:DolphinEndpointSlices,DolphinIdentity:DolphinIdentities,DolphinNode:DolphinNodes,DolphinGatewayClassConfig:DolphinGatewayClassConfigs"

# Generate deepcopy methods
kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/k8s/apis"

# Generate client, informers, listers
kube::codegen::gen_client \
    --input-pkg-root github.com/ccfish2/infra/pkg/k8s/apis \
    --output-pkg-root github.com/ccfish2/infra/pkg/k8s/client \
    --output-base "${TMPDIR}" \
    --with-watch \
    --plural-exceptions ${PLURAL_EXCEPTIONS} \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt" \
    ./pkg/k8s/apis

cp -r "${TMPDIR}/github.com/ccfish2/infra/." ./