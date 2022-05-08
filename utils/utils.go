package utils

import (
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetPublicIP() (ip string, err error) {
	var url = "http://ifconfig.io/ip"
	response, err := http.Get(url)
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
