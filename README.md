# Home-ddns

Simple Golang tool to update DNS records (extensible to multiple providers) for home public IP. Emulate Dynamic DNS for free.

## Installation

The tool can be easily compiled locally:

```bash
git clone https://github.com/Sudneo/Home-ddns.git
cd Home-ddns
make build
./home-ddns -h
```

Alternatively, a prebuild release can be downloaded from [Github](https://github.com/Sudneo/Home-ddns/releases).

### Docker

A Dockerfile is provided to build the tool using Docker:

```bash
cd Home-ddns
docker build -t home-ddns .
docker run -v ${PWD}/config.yaml:/home-ddns/config.yaml -it home-ddns [-v -j -cron -interval 1]
```

Alternative it's possible to use the Makefile:

```bash
make docker-build
make docker-run
```

The image is built using a `scratch` container, so it's very minimal and does not even have a shell (hence, you won't be able to `exec` inside it).

```
home-ddns      latest          2a04c146665c   About a minute ago   5.48MB
```

## Usage 

All the configuration is in a YAML file, this allows to manage multiple providers/domains/records at once, with full flexibility in each domain type, TTL, etc.

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

The usage therefore now is simple:

```
Usage of ./home-ddns:
  -config string
        Configuration file to use (default "config.yaml")
  -cron
        Enable cron mode (execute every interval)
  -interval int
        Interval in minutes between each execution (requires cron mode) (default 60)
  -j    Enable logging in JSON
  -v    Enable debug logs
```

The `cron` mode simply will have the execution run in an infinite loop. At every loop the configuration is re-read, so it can be modified dynamically (for example as a ConfigMap in Kubernetes).
        
## Use Case

The main use case is to update a given DNS record (or a set of record) for a domain in a DNS provider with the public IP address of the machine where this script is run from.
This is the case for example for non-static domestic IPs, where we want to achieve Dynamic-DNS to have our home IPs publicly reachable with a DNS name.
However, the tool supports specifying a value for a domain record, if this is different from the public IP, and can be used to set whetever record type the provider supports.

Note:
The script uses `http://ifconfig.io` to query the public IP. This means that the folks running that site will be able to see your IP address (as you are making a request to them). If this is not fine for you, consider selfhosting a similar service.

## Features

* The record to be set is first queried, if it exists and matches the specified value (or alternatively the public IP), nothing is done.
* If the record exists but doesn't match the current public IP/the specified value, that record is updated.
* If the record does not exist, it is created.
* Cron mode, Docker friendly way to run the tool periodically without having to install cron inside the image.

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

var providersMap = map[string]models.Provider{
	godaddyProvider: &api.GodaddyHandler{},
	myProvider: &api.MyProviderHandler{},
}
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

