package utils

import (
	"log"
	"regexp"
	"strings"
)

func NewHostUtils() HostUtils {
	return &hostUtils{}
}

type HostUtils interface {
	ParseHostName(hostName string) string
}

type hostUtils struct{}

// Parse the string and make it compliant with RFC 1123 host names, by removing invalid characters
func (r *hostUtils) ParseHostName(name string) string {
	newName := strings.ToLower(name)
	invalidCharacters := []string{":", ";", "_"}

	for _, invalidString := range invalidCharacters {
		newName = strings.Replace(newName, invalidString, "-", -1)
	}

	_, err := regexp.MatchString("[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*", newName)

	if err != nil {
		log.Fatal("Error matching host name:", err.Error())
	}

	return newName
}
