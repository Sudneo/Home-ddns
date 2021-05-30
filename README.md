# GOdaddy-ddns
Simple Golang script to update DNS records in godaddy for home public IP 

## Usage 

```
Usage of ./Godaddy-dns:
  -clientKey string
    	Godaddy client key
  -clientid string
    	Godaddy client ID
  -domain string
    	The Godaddy domain to use
  -name string
    	The name of the record, aka the hostname
  -type string
    	The DNS record type to use (default "A")
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
