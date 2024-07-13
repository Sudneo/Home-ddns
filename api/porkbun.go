package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/sudneo/home-ddns/models"
)

const (
	porkbunBaseURL = "https://api.porkbun.com"
)

type PorkbunHandler struct {
	ClientID  string
	ClientKey string
}

type porkbunErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Structure for a DNS record in Porkbun
type porkbunRecordData struct {
	Status  string `json:"status"`
	Records []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		Value    string `json:"content"`
		TTL      string `json:"ttl"`
		Priority int    `json:"priority"`
		Notes    string `json:"notes"`
	}
}

type porkbunAuthData struct {
	ApiKey       string `json:"apikey"`
	SecretApiKey string `json:"secretapikey"`
}

type porkbunUpdateRecordData struct {
	ApiKey       string `json:"apikey"`
	SecretApiKey string `json:"secretapikey"`
	Content      string `json:"content"`
	TTL          string `json:"ttl"`
	Prio         int    `json:"prio"`
}

type porkbunCreateRecordData struct {
	ApiKey       string `json:"apikey"`
	SecretApiKey string `json:"secretapikey"`
	Name         string `json:"name"`
	Recordtype   string `json:"type"`
	Content      string `json:"content"`
	TTL          string `json:"ttl"`
	Prio         int    `json:"prio"`
}

func (h *PorkbunHandler) SetAPIKey(key string) error {
	h.ClientKey = key
	return nil
}

func (h *PorkbunHandler) SetAPIID(id string) error {
	h.ClientID = id
	return nil
}

// GetRecord implements Provider.GetRecord. Fetches from Porkbun API the information about an existing record
func (h *PorkbunHandler) GetRecord(domain string, record models.DNSRecord) (dnsRecord models.DNSRecord, err error) {
	var d models.DNSRecord
	data := porkbunAuthData{
		ApiKey:       h.ClientID,
		SecretApiKey: h.ClientKey,
	}
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return d, err
	}
	url := fmt.Sprintf("%s/api/json/v3/dns/retrieveByNameType/%s/%s/%s", porkbunBaseURL, domain, record.Type, record.Name)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return d, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		response := porkbunErrorResponse{}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return d, err
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			return dnsRecord, err
		}
		return d, &ErrAPIFailed{Code: response.Status, Message: response.Message}
	}
	response := porkbunRecordData{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return d, err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return d, err
	}
	if len(response.Records) == 0 {
		return d, nil
	}
	d.Name = response.Records[0].Name
	d.Value = response.Records[0].Value
	d.Type = response.Records[0].Type
	ttl, err := strconv.Atoi(response.Records[0].TTL)
	if err != nil {
		d.TTL = 0
	}
	d.TTL = ttl
	return d, nil
}

// UpdateRecord implements Provider.UpdateRecord. Updates an existing DNS record with a new configuration
// Generally, this method is invoked when the IP changed
func (h *PorkbunHandler) UpdateRecord(domain string, record models.DNSRecord) (err error) {
	url := fmt.Sprintf("%s/api/json/v3/dns/editByNameType/%s/%s/%s", porkbunBaseURL, domain, record.Type, record.Name)
	data := porkbunUpdateRecordData{
		ApiKey:       h.ClientID,
		SecretApiKey: h.ClientKey,
		Content:      record.Value,
	}
	if record.TTL != 0 && record.TTL < 3600 {
		data.TTL = "3600"
	} else {
		data.TTL = fmt.Sprintf("%d", record.TTL)
	}
	if record.Priority != 0 {
		data.Prio = record.Priority
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
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

	body, err := io.ReadAll(resp.Body)
	fmt.Print(string(body))
	if err != nil {
		log.Fatal(err)
	}

	response := porkbunErrorResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	return &ErrAPIFailed{Code: response.Status, Message: response.Message}
}

// SetRecord implements Provider.SetRecord. Creates a new DNS record as passed in parameters
func (h *PorkbunHandler) SetRecord(domain string, record models.DNSRecord) (err error) {
	url := fmt.Sprintf("%s/api/json/v3/dns/create/%s", porkbunBaseURL, domain)
	data := porkbunCreateRecordData{
		SecretApiKey: h.ClientKey,
		ApiKey:       h.ClientID,
		Name:         record.Name,
		Recordtype:   record.Type,
		Content:      record.Value,
	}

	if record.TTL != 0 && record.TTL < 3600 {
		data.TTL = "3600"
	} else {
		data.TTL = string(record.TTL)
	}

	if record.Priority != 0 {
		data.Prio = record.Priority
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
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

	body, err := io.ReadAll(resp.Body)
	fmt.Print(string(body))
	if err != nil {
		log.Fatal(err)
	}

	response := porkbunErrorResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Print(string(body))
		return err
	}
	return &ErrAPIFailed{Code: response.Status, Message: response.Message}
}
