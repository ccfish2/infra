package config

import (
	"io"

	"github.com/ccfish2/infra/pkg/safeio"
	metallbcfg "github.com/ccfish2/metalb0110/pkg/config"
)

func Parse(r io.Reader) (*metallbcfg.Config, error) {
	buf, err := safeio.ReadAllLimit(r, safeio.MB) // 1MB
	if err != nil {
		return nil, err
	}
	cfg, err := metallbcfg.Parse(buf)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
