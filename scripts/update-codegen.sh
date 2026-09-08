#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/.."
GOPATH=${GOPATH:-$(go env GOPATH)}
CODEGEN_PKG="${GOPATH}/pkg/mod/k8s.io/code-generator@v0.35.6"

source "${CODEGEN_PKG}/kube_codegen.sh"

TMPDIR="${TMPDIR:-/tmp/codegen-$(date +%s)}"
mkdir -p "${TMPDIR}"

PLURAL_EXCEPTIONS="DolphinEndpoints:DolphinEndpoints,DolphinEnvoyConfig:DolphinEnvoyConfigs,DolphinEndpointSlice:DolphinEndpointSlices,DolphinIdentity:DolphinIdentities,DolphinNode:DolphinNodes,DolphinGatewayClassConfig:DolphinGatewayClassConfigs"

# Generate deepcopy and deepequal methods
kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/k8s/apis"

# Generate client, informers, and listers
kube::codegen::gen_client \
    "./pkg/k8s/apis" \
    --with-watch \
    --output-dir "${TMPDIR}/github.com/ccfish2/infra/pkg/k8s/client" \
    --output-pkg "github.com/ccfish2/infra/pkg/k8s/client" \
    --plural-exceptions ${PLURAL_EXCEPTIONS} \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt"

# Remove old generated client files (use sudo if needed)
sudo rm -rf "${SCRIPT_ROOT}/pkg/k8s/client/clientset" \
            "${SCRIPT_ROOT}/pkg/k8s/client/informers" \
            "${SCRIPT_ROOT}/pkg/k8s/client/listers"

# Copy generated files back to project
cp -r "${TMPDIR}/github.com/ccfish2/infra/pkg/k8s/client/." "${SCRIPT_ROOT}/pkg/k8s/client/"

echo "Code generation complete!"