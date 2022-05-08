# GOdaddy-ddns
Simple Golang script to update DNS records in godaddy for home public IP 

## Usage 

All the configuration is now moved to a YAML file, this allows to manage multiple domains at once, with more flexibility in record types,TTLs, etc.

An example configuration is as follows:

```yaml
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
Usage of ./godaddy-dns:
  -config string
          Configuration file to use (default "config.yaml")
  -v    Enable debug logs
```
	    
## Use Case

The main use case is to update a given DNS record for a domain in Godaddy with the public IP address of the machine where this script is run from.
This is the case for example for non-static domestic IPs, where we want to achieve Dynamic-DNS to have our home IPs publicly reacheable with a DNS name.

Note:
The script uses `http://ifconfig.io` to query the public IP. This means that the folks running that site will be able to see your IP address (as you are making a request to them). If this is not fine for you, consider selfhosting the ifconfig.io (I might do that in the future and parametrize the respective function).

## Features

* The record to be set is first queried, if it exists and matches the current public IP, nothing is done.
* If the record exists but doesn't match the current public IP, that record is updated.
* If the record does not exist, it is created.

## Roadmap

I am planning and laid the foundation to allow different providers to be 'plugged in'. This means that it will be enough for someone to write the API interactions with their own provider, and the rest of the logic will be shared with the rest of the code.
