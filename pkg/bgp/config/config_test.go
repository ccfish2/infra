package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const yaml = `---
peers:
  - peer-address: 172.16.0.4
    peer-asn: 75612
    my-asn: 75612
address-pools:
  - name: dolphin
    protocol: bgp
    addresses:
      - 172.16.0.150/29
`

var json = `{"peers":[{"peer-address":"172.16.0.4","peer-asn":75612,"my-asn":75612}],"address-pools":[{"name":"default","protocol":"bgp","addresses":["172.16.0.150/29"]}]}`

func Test_Parse(t *testing.T) {
	_, err := Parse(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("failed to parse yaml: %v", err)
	}

	_, err = Parse(strings.NewReader(json))
	if err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	cfg, err := Parse(strings.NewReader(`{"json":"random"}`))
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "failed to read MetalLB config:"))
	assert.Nil(t, cfg)

}
