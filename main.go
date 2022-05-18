package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	godaddy "github.com/sudneo/godaddy-dns/api"
	"github.com/sudneo/godaddy-dns/config"
	"github.com/sudneo/godaddy-dns/models"
	"github.com/sudneo/godaddy-dns/utils"
)

const (
	godaddyProvider = "Godaddy"
)

var providersMap = map[string]models.Provider{
	godaddyProvider: &godaddy.GodaddyHandler{},
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func processDomain(d config.DomainConfiguration, handler models.Provider, externalIP string) error {
	for _, record := range d.Records {
		dnsRecord, err := handler.GetRecord(d.Domain, record)
		if err != nil {
			log.Error(err)
			continue
		}
		if record.Value == "" {
			if record.Type == "CNAME" {
				record.Value = "@"
			} else {
				record.Value = externalIP
			}
		}
		if dnsRecord.Value == "" {
			log.WithFields(log.Fields{
				"Name": record.Name,
			}).Debug("Not found existing record for domain, creating a new one")
			err = handler.SetRecord(d.Domain, record)
		} else {
			if dnsRecord.Value != record.Value {
				log.WithFields(log.Fields{
					"Name": record.Name,
				}).Debug("Existing record found with old data, updating")
				err = handler.UpdateRecord(d.Domain, record)
			} else {
				log.WithFields(log.Fields{
					"Name":  record.Name,
					"Value": record.Value,
					"DNS":   dnsRecord.Value,
				}).Debug("Correct record already exists, nothing to do")
			}
		}
	}
	return nil
}

func run(c config.Config) error {
	externalIP, err := utils.GetPublicIP()
	if err != nil {
		return err
	}
	// Process providers one by one
	for _, provider := range c.Providers {
		// Match the provider name with the corresponding type using the global map
		handler, ok := providersMap[provider.Name]
		if ok {
			handler.SetAPIID(provider.ClientID)
			handler.SetAPIKey(provider.ClientKey)
			log.WithFields(log.Fields{
				"Domains":  len(provider.Domains),
				"Provider": provider.Name,
			}).Debug("Processing domains")
			for _, domain := range provider.Domains {
				err := processDomain(domain, handler, externalIP)
				if err != nil {
					log.Error(err)
				}
			}

		} else {
			log.WithFields(log.Fields{
				"Provider": provider.Name,
			}).Error("Provider not recognized")
		}
	}

	return nil
}

func main() {
	var configuration = flag.String("config", "config.yaml", "Configuration file to use")
	var debug = flag.Bool("v", false, "Enable debug logs")
	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	conf, err := config.ReadConfig(*configuration)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = run(conf)
	if err != nil {
		log.Error(err)
	}
}
