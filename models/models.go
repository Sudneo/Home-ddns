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
