#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)/.."
GOPATH=${GOPATH:-$(go env GOPATH)}
CODEGEN_PKG="${GOPATH}/pkg/mod/k8s.io/code-generator@v0.35.6"

source "${CODEGEN_PKG}/kube_codegen.sh"

TMPDIR="${TMPDIR:-/tmp/codegen-v2alpha1-$(date +%s)}"
mkdir -p "${TMPDIR}"

PLURAL_EXCEPTIONS="DolphinGatewayClassConfig:DolphinGatewayClassConfigs"

# Generate deepcopy for v2alpha1 only
kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/k8s/apis/dolphin.io/v2alpha1"

# Temporarily hide v1
mv "${SCRIPT_ROOT}/pkg/k8s/apis/dolphin.io/v1" "${SCRIPT_ROOT}/pkg/k8s/apis/dolphin.io/v1.bak"

# Generate client - pass just the APIs root
kube::codegen::gen_client \
    "./pkg/k8s/apis" \
    --with-watch \
    --output-dir "${TMPDIR}/github.com/ccfish2/infra/pkg/k8s/client" \
    --output-pkg "github.com/ccfish2/infra/pkg/k8s/client" \
    --plural-exceptions ${PLURAL_EXCEPTIONS} \
    --boilerplate "${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt"

# Restore v1
mv "${SCRIPT_ROOT}/pkg/k8s/apis/dolphin.io/v1.bak" "${SCRIPT_ROOT}/pkg/k8s/apis/dolphin.io/v1"

cp -r "${TMPDIR}/github.com/ccfish2/infra/pkg/k8s/client/." "${SCRIPT_ROOT}/pkg/k8s/client/"

echo "Code generation for v2alpha1 complete!"