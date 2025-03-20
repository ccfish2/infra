package config

const yaml (
	`---
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
	json = `{"peers":[{"peer-address":"172.16.0.4","peer-asn":75612,"my-asn":75612}],"address-pools":[{"name":"default","protocol":"bgp","addresses":["172.16.0.150/29"]}]}`
)

func Test_Parse(t *testing.T) {
	cfg, err := Parse(strings.NewReader(yaml))
	if err != nil || cfg ==nil {
		t.Fatalf("failed to parse yaml: %v", err)
	}

	ParseJSON, err := Parse(strings.NewReader(json))
	if err != nil || ParseJSON == nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	parsedJ, err := Parse(strings.NewReader(`{"JSON":"RANDOM"}`))
	if err != nil || parsedJ != nil {
		t.Fatalf("expected error for invalid JSON, got: %v", err)
	}
}