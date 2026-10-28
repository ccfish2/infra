package crd

import (
	_ "embed"
)

// DolphinEnvoyConfigs contains the generated DolphinEnvoyConfig CRD.
//
//go:embed bases/dolphinenvoyconfigs.yaml
var DolphinEnvoyConfigs []byte
