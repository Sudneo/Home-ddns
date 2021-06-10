package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	APIBaseURL = "https://api.godaddy.com"
)

type setRecordData struct {
	Data     string `json:"data"`
	Port     int    `json:"port"`
	Priority int    `json:"priority"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	TTL      int    `json:"ttl"`
	Weight   int    `json:"weight"`
}

// Sometimes Godaddy API returns code as a string, sometimes as an int
// ¯\_(ツ)_/¯

type APIErrorResponseInt struct {
	Code   int `json:"code"`
	Fields []struct {
		Code        int    `json:"code"`
		Message     string `json:"message"`
		Path        string `json:"path"`
		Pathrelated string `json:"pathRelated"`
	} `json:"fields"`
	Message string `json:"message"`
}

type APIErrorResponse struct {
	Code   string `json:"code"`
	Fields []struct {
		Code        string `json:"code"`
		Message     string `json:"message"`
		Path        string `json:"path"`
		Pathrelated string `json:"pathRelated"`
	} `json:"fields"`
	Message string `json:"message"`
}

// Structure for a DNS record in GOdaddy
type RecordData []struct {
	Data     string `json:"data"`
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Priority int    `json:"priority"`
	Protocol string `json:"protocol"`
	Service  string `json:"service"`
	TTL      int    `json:"ttl"`
	Type     string `json:"type"`
	Weight   int    `json:"weight"`
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func getPublicIP() (ip string, err error) {
	var url = "http://ifconfig.io/ip"
	response, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return "", err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		return "", err
	}
	return strings.TrimSuffix(string(responseData), "\n"), nil
}

func updateDNSRecord(domain string, name string, recordType string, clientID string, clientKey string, value string) (err error) {
	url := fmt.Sprintf("%s/v1/domains/%s/records/%s/%s", APIBaseURL, domain, recordType, name)
	// Note: the documentation states that 'name' and 'type' are not needed (as they are in URL already)
	// but API calls fail without.
	data := RecordData{
		{
			Data:     value,
			Name:     name,
			Type:     recordType,
			Port:     1,
			Priority: 0,
			Protocol: "",
			Service:  "",
			TTL:      600,
			Weight:   0,
		}}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	authHeader := fmt.Sprintf("sso-key %s:%s", clientID, clientKey)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode == 200 {
		log.Info("Successfully updated DNS record")
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Print(string(body))
	if err != nil {
		log.Fatal(err)
	}
	response := APIErrorResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"Code":    response.Code,
		"Message": response.Message,
	}).Error("Failed to perform PUT request to Godaddy API")

	return fmt.Errorf("Failed to update DNS record")
}

func addDNSRecord(domain string, name string, recordType string, clientID string, clientKey string, value string) (err error) {
	url := fmt.Sprintf("%s/v1/domains/%s/records", APIBaseURL, domain)
	data := RecordData{
		{
			Data:     value,
			Name:     name,
			Type:     recordType,
			Port:     1,
			Priority: 0,
			Protocol: "",
			Service:  "",
			TTL:      600,
			Weight:   0,
		}}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	authHeader := fmt.Sprintf("sso-key %s:%s", clientID, clientKey)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode == 200 {
		log.Info("Successfully created DNS record")
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	fmt.Print(string(body))
	if err != nil {
		log.Fatal(err)
	}
	response := APIErrorResponseInt{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"Code":    response.Code,
		"Message": response.Message,
	}).Error("Failed to perform PATCH request to Godaddy API")

	return fmt.Errorf("Failed to create DNS record")
}

func getCurrentRecord(domain string, name string, recordType string, clientID string, clientKey string) (currentIP string, err error) {

	url := fmt.Sprintf("%s/v1/domains/%s/records/%s/%s", APIBaseURL, domain, recordType, name)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	authHeader := fmt.Sprintf("sso-key %s:%s", clientID, clientKey)
	req.Header.Add("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		log.Error("Request failed")
		response := APIErrorResponse{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			return "", err
		}
		log.WithFields(log.Fields{
			"Code":    response.Code,
			"Message": response.Message,
		}).Error("Failed to perform GET request to Godaddy API")
	} else {
		response := RecordData{}
		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return "", err
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			return "", err
		}
		if len(response) == 0 {
			log.Error("No record with this name found")
			return "", nil
		} else {
			log.WithFields(log.Fields{
				"IP":   response[0].Data,
				"Type": response[0].Type,
				"TTL":  response[0].TTL,
				"Name": response[0].Name,
			}).Info("Found existing record")
			return response[0].Data, nil
		}
	}
	return "", nil
}

func main() {

	var domain = flag.String("domain", "", "The Godaddy domain to use")
	// var name = flag.String("name", "", "The name of the record, aka the hostname")
	var recordType = flag.String("type", "A", "The DNS record type to use")
	var clientID = flag.String("clientid", "", "Godaddy client ID")
	var clientKey = flag.String("clientKey", "", "Godaddy client key")
	flag.Parse()
	names := flag.Args()
	if *domain == "" || *clientID == "" || *clientKey == "" {
		log.Fatal("Please supply all the necessary arguments")
	}
	publicIP, err := getPublicIP()
	if err != nil {
		log.Fatal("Failed to get public IP")
	}
	log.WithFields(log.Fields{
		"IP": publicIP,
	}).Info("Found public IP")
	for _, name := range names {
		currentIP, err := getCurrentRecord(*domain, name, *recordType, *clientID, *clientKey)
		if err != nil {
			log.Fatal("Failed to get current record")
		}
		if currentIP == "" {
			log.Info("The record currently does not exist, will be created")
			err = addDNSRecord(*domain, name, *recordType, *clientID, *clientKey, publicIP)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			if currentIP == publicIP {
				log.Info("The record is already up-to-date with current public IP")
				continue
			} else {
				log.Info("The DNS record exists, but needs to be updated with current IP")
				err = updateDNSRecord(*domain, name, *recordType, *clientID, *clientKey, publicIP)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
