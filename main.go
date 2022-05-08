package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	godaddy "github.com/sudneo/godaddy-dns/api"
	"github.com/sudneo/godaddy-dns/config"
	"github.com/sudneo/godaddy-dns/utils"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func processDomain(d config.DomainConfiguration, handler godaddy.GodaddyHandler, externalIP string) error {
	for _, record := range d.Records {
		dnsRecord, err := handler.GetRecord(d.Domain, record)
		if err != nil {
			log.Error(err)
			continue
		}
		if record.Type == "CNAME" {
			if record.Value == "" {
				record.Value = "@"
			}
		} else {
			record.Value = externalIP
		}
		if dnsRecord.Value == "" {
			log.WithFields(log.Fields{
				"Name": record.Name,
			}).Debug("Not found existing record for domain, creating a new one")
			err = handler.SetRecord(d.Domain, record)
		} else {
			if dnsRecord.Value != externalIP {
				log.WithFields(log.Fields{
					"Name": record.Name,
				}).Debug("Existing record found with old data, updating")
				err = handler.UpdateRecord(d.Domain, record)
			}
		}
	}
	return nil
}

func processDomains(c config.Config) error {
	externalIP, err := utils.GetPublicIP()
	if err != nil {
		return err
	}
	handler := godaddy.GodaddyHandler{
		ClientID:  c.ClientID,
		ClientKey: c.ClientKey,
	}
	log.WithFields(log.Fields{
		"Domains": len(c.Domains),
	}).Debug("Processing domains")
	for _, domain := range c.Domains {
		err := processDomain(domain, handler, externalIP)
		if err != nil {
			log.Error(err)
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
	err = processDomains(conf)
	if err != nil {
		log.Error(err)
	}
}
