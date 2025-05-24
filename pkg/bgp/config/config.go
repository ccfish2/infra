package config

import (
	"fmt"
	"io"

	"github.com/ccfish2/infra/pkg/safeio"
	metallbcfg "github.com/ccfish2/metalb0110/pkg/config"
)

func Parse(r io.Reader) (*metallbcfg.Config, error) {
	buf, err := safeio.ReadAllLimit(r, safeio.MB) // 1MB
	if err != nil {
		return nil, fmt.Errorf("failed to read MetalLB config: %w", err)
	}
	cfg, err := metallbcfg.Parse(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read MetalLB config: %w", err)
	}
	return cfg, nil
}
