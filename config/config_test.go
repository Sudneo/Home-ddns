package config

import (
	"testing"
)

var validConfig = []byte(`
providers:
  - name: provider1
    client_id: "id"
    client_key: "key"
    domains:
      - domain: example.com
        records:
          - name: test
            type: A
            ttl: 70
          - name: ctest
            type: CNAME
`)

var missingKeyConfig = []byte(`
providers:
  - name: provider1
    client_id: "id"
    domains:
      - domain: example.com
        records:
          - name: test
            type: A
            ttl: 70
          - name: ctest
            type: CNAME
`)

var complexConfig = []byte(`
providers:
  - name: provider1
    client_id: "id"
    client_key: "key"
    domains:
      - domain: example.com
        records:
          - name: test
            type: A
            ttl: 70
          - name: ctest
            type: CNAME
  - name: provider2
    client_id: "id"
    client_key: "key"
    domains:
      - domain: test.com
        records:
          - name: home
            type: A
          - name: vpn
            type: CNAME
            value: home
      - domain: home.net
        records:
          - name: mail
            type: MX
`)

func TestParseConfig(t *testing.T) {

	config, err := parseConfig(validConfig)
	if err != nil {
		t.Errorf("Parsing the configuration YAML lead to error")
	}
	expectedRecords := 2
	actualRecords := len(config.Providers[0].Domains[0].Records)
	if actualRecords != expectedRecords {
		t.Errorf("Configuration not parsed correctly. Expected %d records, found %d", expectedRecords, actualRecords)
	}
	config, err = parseConfig(missingKeyConfig)
	if err == nil {
		t.Errorf("Invalid configuration did not error, API key is missing")
	}
	config, err = parseConfig(complexConfig)
	if err != nil {
		t.Errorf("Parsing the complex YAML lead to error: %s", err)
	}
	expectedProviders := 2
	actualProviders := len(config.Providers)
	if expectedProviders != actualProviders {
		t.Errorf("Complex config not parsed correctly: expected %d providers, found %d", expectedProviders, actualProviders)
	}
	expectedDomains := 2
	actualDomains := len(config.Providers[1].Domains)
	if expectedDomains != actualDomains {
		t.Errorf("Complex config not parsed correctly: expected %d domains, found %d", expectedDomains, actualDomains)
	}
	if config.Providers[0].Domains[0].Records[1].Name != "ctest" {
		t.Errorf("Failed to parse multiple records for a domain in a provider")
	}
	if config.Providers[1].Domains[1].Records[0].Type != "MX" {
		t.Errorf("Failed to parse record in multiple domains for a second provider")
	}
}

func TestReadConfig(t *testing.T) {
	filename := "../test/config.yaml"
	_, err := ReadConfig(filename)
	if err != nil {
		t.Errorf("Reading the configuration YAML lead to error")
	}
}
