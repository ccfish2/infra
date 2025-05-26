//go:build tools
// +build tools

package tools

import (
	_ "github.com/cilium/deepequal-gen"

	_ "k8s.io/code-generator"
	_ "k8s.io/code-generator/cmd/client-gen"

	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
