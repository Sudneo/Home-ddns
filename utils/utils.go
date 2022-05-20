package utils

import (
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	ifconfigURL = "http://ifconfig.io/ip"
)

func GetPublicIP() (ip string, err error) {
	response, err := http.Get(ifconfigURL)
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
