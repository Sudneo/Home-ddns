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
	if len(config.Domains) == 0 {
		return config, &InvalidConfiguration{Description: "No domain configuration"}
	}
	if config.ClientID == "" || config.ClientKey == "" {
		return config, &InvalidConfiguration{Description: "No API credentials supplied"}
	}
	return config, nil

}
