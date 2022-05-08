package api

import (
	"fmt"

	"github.com/sudneo/godaddy-dns/models"
)

// Generic interface for a provider
type Provider interface {
	// Given a record, determine current value
	GetRecord(domain string, record models.DNSRecord) (models.DNSRecord, error)
	// Create a new record for a host.domain
	SetRecord(domain string, record models.DNSRecord) error
	// Update an existing record for a host.domain
	UpdateRecord(domain string, record models.DNSRecord) error
}

type ErrAPIFailed struct {
	Message string
	Code    string
}

func (e *ErrAPIFailed) Error() string {
	return fmt.Sprintf("API call failed with code %s: %s", e.Code, e.Message)
}
