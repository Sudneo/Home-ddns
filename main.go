package main

import (
	"errors"
	"flag"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/sudneo/home-ddns/api"
	"github.com/sudneo/home-ddns/config"
	"github.com/sudneo/home-ddns/models"
	"github.com/sudneo/home-ddns/utils"
)

const (
	godaddyProvider = "Godaddy"
)

// Map to register providers
// Each name (used in the config) is matched
// with the corresponding handler type
var providersMap = map[string]models.Provider{
	godaddyProvider: &api.GodaddyHandler{},
}

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func processDomain(d config.DomainConfiguration, handler models.Provider, externalIP string) error {
	for _, record := range d.Records {
		dnsRecord, err := handler.GetRecord(d.Domain, record)
		if err != nil {
			log.WithFields(log.Fields{
				"Error":  err,
				"Record": record.Name,
			}).Error("Failed to process DNS record")
			continue
		}
		// If the DNS record does not have a value specified, set sane defaults
		if record.Value == "" {
			if record.Type == "CNAME" {
				record.Value = "@"
			} else {
				record.Value = externalIP
			}
		}
		// If the current record does not exist, the DNS record must be created
		if dnsRecord.Value == "" {
			log.WithFields(log.Fields{
				"Name": record.Name,
			}).Debug("Not found existing record for domain, creating a new one")
			err = handler.SetRecord(d.Domain, record)
		} else {
			// If the record does exist, but it's not up-to-date, update it
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
	// This call is done here to minimize requests to third parties
	externalIP, err := utils.GetPublicIP()
	if err != nil {
		return err
	}
	if len(externalIP) == 0 {
		log.Error("No external IP obtained, likely failed to call ifconfig.io")
		return errors.New("Failed to query external IP")
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
			}).Debug("Processing domains for provider")
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
			continue
		}
	}
	return nil
}

func main() {
	var configuration = flag.String("config", "config.yaml", "Configuration file to use")
	var debug = flag.Bool("v", false, "Enable debug logs")
	var json = flag.Bool("j", false, "Enable logging in JSON")
	var cronMode = flag.Bool("cron", false, "Enable cron mode (execute every interval)")
	var cronInterval = flag.Int("interval", 60, "Interval in minutes between each execution (requires cron mode)")
	flag.Parse()
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	if *json {
		log.SetFormatter(&log.JSONFormatter{})
	}
	if !*cronMode {
		conf, err := config.ReadConfig(*configuration)
		if err != nil {
			log.Fatal(err)
			return
		}
		err = run(conf)
		if err != nil {
			log.Error(err)
		}
		log.Info("Processing completed successfully")
	} else {
		for {
			conf, err := config.ReadConfig(*configuration)
			if err != nil {
				log.Fatal(err)
				return
			}
			log.Debug("Configuration reloaded")
			err = run(conf)
			if err != nil {
				log.Error(err)
			}
			log.Debug("Execution completed successfully")
			time.Sleep(time.Duration(*cronInterval) * time.Minute)
		}
	}
}
