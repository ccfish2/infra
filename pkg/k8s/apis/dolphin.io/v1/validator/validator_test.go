package validator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDECValidation(t *testing.T) {
	decValidator, err := NewDECValidator()
	require.NoError(t, err)

	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid DEC",
			yaml: `apiVersion: dolphin.io/v1
kind: DolphinEnvoyConfig
metadata:
  name: example-dec
spec:
  services:
  - name: frontend
    namespace: default
    ports: [80, 443]
  backendServices:
  - name: frontend-backend
    namespace: default
    number: ["80", "443"]
  resources:
  - "@type": type.googleapis.com/envoy.config.listener.v3.Listener
    name: example-listener
`,
		},
		{
			name: "missing required service name",
			yaml: `apiVersion: dolphin.io/v1
kind: DolphinEnvoyConfig
metadata:
  name: invalid-dec
spec:
  services:
  - namespace: default
    ports: [80]
`,
			wantErr: true,
		},
		{
			name: "missing required backend service name",
			yaml: `apiVersion: dolphin.io/v1
kind: DolphinEnvoyConfig
metadata:
  name: invalid-dec
spec:
  backendServices:
  - namespace: default
    number: ["80"]
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec, err := DecodeDEC([]byte(tt.yaml))
			require.NoError(t, err)

			err = decValidator.ValidateDEC(dec)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
