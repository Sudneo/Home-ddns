# Home-ddns

Simple Golang tool to update DNS records (extensible to multiple providers) for home public IP 

## Installation

The tool can be easily compiled locally:

```bash
git clone https://github.com/Sudneo/Home-ddns.git
cd home-ddns
make build
```

Alternatively, a prebuild release can be downloaded from [Github](https://github.com/Sudneo/Home-ddns/releases).

## Usage 

All the configuration is in a YAML file, this allows to manage multiple domains at once, with more flexibility in record types, TTLs, etc.

An example configuration is as follows:

```yaml
providers:
  - name: "Godaddy" # The only implemented at the moment
    client_id: "MYID" 
    client_key: "MYKEY"
    domains:
      - domain: "mydomain.com"
        records:
          - name:  "home"
            type:  "A"
          - name:  "test"
            type:  "CNAME"
          - name:  "anotherCNAME"
            type:  "CNAME"
            ttl:   "1200"
          - name:  "proxy"
            type:  "A"
            value: "260.1.1.1"
```

The usage therefore now is much simpler:

```
Usage of ./home-ddns:
  -config string
        Configuration file to use (default "config.yaml")
  -j    Enable logging in JSON
  -v    Enable debug logs
```
        
## Use Case

The main use case is to update a given DNS record for a domain in a DNS provider with the public IP address of the machine where this script is run from.
This is the case for example for non-static domestic IPs, where we want to achieve Dynamic-DNS to have our home IPs publicly reacheable with a DNS name.
However, the tool supports specifying a value for a domain, if this is different from the public IP, and can be used to set whetever record type the provider supports.

Note:
The script uses `http://ifconfig.io` to query the public IP. This means that the folks running that site will be able to see your IP address (as you are making a request to them). If this is not fine for you, consider selfhosting the ifconfig.io (I might do that in the future and parametrize the respective function).

## Features

* The record to be set is first queried, if it exists and matches the current public IP, nothing is done.
* If the record exists but doesn't match the current public IP, that record is updated.
* If the record does not exist, it is created.

## Development

The most common use case for development is adding a new provider to support.
To do this, it's enough to implement the `models.Provider` interface and then register the new provider in the global map in `main.go` to map a string of choice to the corresponding type.

Currently, the only implemented provider is Godaddy:

```go
const (
	godaddyProvider = "Godaddy"
	// New provider here
	myProvider = "MyProvider"
)
```

and 

```go
package api

[...]

type MyProviderHandler struct {
	ClientID  string
	ClientKey string
	MyKey string
}
func (h *MyProviderHandler) SetAPIKey(key string) error {}
func (h *MyProviderHandler) SetAPIID(key string) error {}
func (h *MyProviderHandler) GetRecord(domain string, record models.DNSRecord) (dnsRecord models.DNSRecord, err error) {}
func (h *MyProviderHandler) SetRecord(domain string, record models.DNSRecord) (err error) {}
func (h *MyProviderHandler) UpdateRecord(domain string, record models.DNSRecord) (err error) {}
```

Finally, the new provider can be used directly in the configuration:

```yaml
[...]
  - name: "MyProvider"
    client_id: "MYID" 
    client_key: "MYKEY"
    domains:
    [...]
```

It's important to consider a few things:

* For CNAME records, the value expected is the A record pointer, such as `@`, rather than the IP, this is used for drift detection. At worst, the tool will set the value everytime, even if it's already correct, if this is not implemented correctly.
* When a DNS record does NOT exist, GetRecord should return `""` as the record value.

