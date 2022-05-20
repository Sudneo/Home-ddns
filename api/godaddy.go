package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/sudneo/home-ddns/models"
)

const (
	godaddyAPIBaseURL = "https://api.godaddy.com"
)

type GodaddyHandler struct {
	ClientID  string
	ClientKey string
}

type godaddyErrorResponse struct {
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
type godaddyRecordData []struct {
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

func (h *GodaddyHandler) SetAPIKey(key string) error {
	h.ClientKey = key
	return nil
}

func (h *GodaddyHandler) SetAPIID(id string) error {
	h.ClientID = id
	return nil
}

// GetRecord implements Provider.GetRecord. Fetches from Godaddy API the information about an existing record
func (h *GodaddyHandler) GetRecord(domain string, record models.DNSRecord) (dnsRecord models.DNSRecord, err error) {
	var d models.DNSRecord
	url := fmt.Sprintf("%s/v1/domains/%s/records/%s/%s", godaddyAPIBaseURL, domain, record.Type, record.Name)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	authHeader := fmt.Sprintf("sso-key %s:%s", h.ClientID, h.ClientKey)
	req.Header.Add("Authorization", authHeader)
	resp, err := client.Do(req)
	if err != nil {
		return d, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		response := godaddyErrorResponse{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return d, err
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			return dnsRecord, err
		}
		return d, &ErrAPIFailed{Code: response.Code, Message: response.Message}
	}
	response := godaddyRecordData{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return d, err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return d, err
	}
	if len(response) == 0 {
		return d, nil
	}
	d.Name = response[0].Name
	d.Value = response[0].Data
	d.Type = response[0].Type
	d.TTL = response[0].TTL
	d.Weight = response[0].Weight
	d.Service = response[0].Service
	d.Protocol = response[0].Protocol
	d.Port = response[0].Port
	return d, nil
}

// SetRecord implements Provider.SetRecord. Creates a new DNS record as passed in parameters
func (h *GodaddyHandler) SetRecord(domain string, record models.DNSRecord) (err error) {
	url := fmt.Sprintf("%s/v1/domains/%s/records", godaddyAPIBaseURL, domain)
	// We need an array because Godaddy API can modify multiple records at once
	data := godaddyRecordData{
		{
			Data: record.Value,
			Name: record.Name,
			Type: record.Type,
		}}
	if record.Port < 1 || record.Port > 65535 {
		data[0].Port = 1
	} else {
		data[0].Port = record.Port
	}
	if record.TTL != 0 && record.TTL >= 600 {
		data[0].TTL = record.TTL
	} else {
		data[0].TTL = 600
	}
	if record.Priority != 0 {
		data[0].Priority = record.Priority
	}
	if record.Protocol != "" {
		data[0].Protocol = record.Protocol
	}
	if record.Service != "" {
		data[0].Service = record.Service
	}
	if record.Weight != 0 {
		data[0].Weight = record.Weight
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	authHeader := fmt.Sprintf("sso-key %s:%s", h.ClientID, h.ClientKey)
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

	response := godaddyErrorResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	return &ErrAPIFailed{Code: response.Code, Message: response.Message}
}

// UpdateRecord implements Provider.UpdateRecord. Updates an existing DNS record with a new configuration
// Generally, this method is invoked when the IP changed
func (h *GodaddyHandler) UpdateRecord(domain string, record models.DNSRecord) (err error) {
	url := fmt.Sprintf("%s/v1/domains/%s/records/%s/%s", godaddyAPIBaseURL, domain, record.Type, record.Name)
	data := godaddyRecordData{
		{
			Data: record.Value,
			Name: record.Name,
			Type: record.Type,
		}}
	if record.Port < 1 || record.Port > 65535 {
		data[0].Port = 1
	} else {
		data[0].Port = record.Port
	}
	if record.TTL != 0 && record.TTL >= 600 {
		data[0].TTL = record.TTL
	} else {
		data[0].TTL = 600
	}
	if record.Priority != 0 {
		data[0].Priority = record.Priority
	}
	if record.Protocol != "" {
		data[0].Protocol = record.Protocol
	}
	if record.Service != "" {
		data[0].Service = record.Service
	}
	if record.Weight != 0 {
		data[0].Weight = record.Weight
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	authHeader := fmt.Sprintf("sso-key %s:%s", h.ClientID, h.ClientKey)
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

	response := godaddyErrorResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	return &ErrAPIFailed{Code: response.Code, Message: response.Message}
}
