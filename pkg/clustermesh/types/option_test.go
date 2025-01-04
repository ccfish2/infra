package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterInfoValidate(t *testing.T) {
	OldClusterMaxID := ClusterIDMax
	defer func() {
		ClusterIDMax = OldClusterMaxID
	}()
	tests := []struct {
		cinfo         ClusterInfo
		WantMCCErr    bool
		WantErr       bool
		WantStaticErr bool
	}{
		{
			cinfo:         ClusterInfo{0, "default", 255},
			WantErr:       false,
			WantStaticErr: true,
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("ID %d Name %s MaxConnectedServer %d", tt.cinfo.Id, tt.cinfo.Name, tt.cinfo.MaxConnectedServer)
		t.Run(name, func(t *testing.T) {
			if tt.WantMCCErr {
				assert.Error(t, tt.cinfo.InitClusterIDMax())
			} else {
				assert.NoError(t, tt.cinfo.InitClusterIDMax())
			}

			if tt.WantErr {
				assert.Error(t, tt.cinfo.Validate())
			} else {
				assert.NoError(t, tt.cinfo.Validate())
			}

			if tt.WantStaticErr {
				assert.Error(t, tt.cinfo.ValidateStrict())
			} else {
				assert.NoError(t, tt.cinfo.ValidateStrict())
			}
		})
	}

}
