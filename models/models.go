package models

type DNSRecord struct {
	Name     string `yaml:"name"`
	Value    string `yaml:"value"`
	Type     string `yaml:"type"`
	TTL      int    `yaml:"ttl"`
	Weight   int    `yaml:"weight"`
	Service  string `yaml:"service"`
	Protocol string `yaml:"protocol"`
	Priority int    `yaml:"priority"`
	Port     int    `yaml:"port"`
}

// Generic interface for a provider
type Provider interface {
	// Given a record, determine current value
	GetRecord(domain string, record DNSRecord) (DNSRecord, error)
	// Create a new record for a host.domain
	SetRecord(domain string, record DNSRecord) error
	// Update an existing record for a host.domain
	UpdateRecord(domain string, record DNSRecord) error
	// Set API key for the given provider
	SetAPIKey(key string) error
	// Set API id for the given provider
	SetAPIID(id string) error
}
