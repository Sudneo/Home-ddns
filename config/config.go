package config

import (
	"fmt"
	"io/ioutil"

	"github.com/sudneo/godaddy-dns/models"
	yaml "gopkg.in/yaml.v3"
)

type InvalidConfiguration struct {
	Description string
}

func (i *InvalidConfiguration) Error() string {
	return fmt.Sprintf("Invalid configuration: %s", i.Description)
}

type Config struct {
	Providers []ProviderConfiguration `yaml:"providers"`
}

type ProviderConfiguration struct {
	Name      string                `yaml:"name"`
	Domains   []DomainConfiguration `yaml:"domains"`
	ClientID  string                `yaml:"client_id"`
	ClientKey string                `yaml:"client_key"`
}

type DomainConfiguration struct {
	Domain  string             `yaml:"domain"`
	Records []models.DNSRecord `yaml:"records"`
}

func ReadConfig(configFile string) (Config, error) {
	var config Config
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return config, &InvalidConfiguration{Description: fmt.Sprintf("Could not read file %s", configFile)}
	}
	config, err = parseConfig(yamlFile)
	return config, err
}

func parseConfig(yamlData []byte) (Config, error) {
	var config Config
	err := yaml.Unmarshal(yamlData, &config)
	if err != nil {
		return config, &InvalidConfiguration{Description: "Invalid YAML in configuration"}
	}
	if len(config.Providers) == 0 {
		return config, &InvalidConfiguration{Description: "No provider configuration supplied"}
	}
	totalDomains := 0
	for _, provider := range config.Providers {
		totalDomains += len(provider.Domains)
		if provider.ClientID == "" || provider.ClientKey == "" {
			return config, &InvalidConfiguration{Description: "Provider configured but no API credentials supplied"}
		}
	}
	if totalDomains == 0 {
		return config, &InvalidConfiguration{Description: "No domain configuration supplied"}
	}
	return config, nil
}
