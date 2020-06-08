package main

import (
	"encoding/json"
	"log"
	"os"
)

// Configuration holds variables required for application's initialization
type Configuration struct {
	ServerCert              string
	ServerCertKey           string
	DBuser                  string
	DBpassword              string
	DBhost                  string
	DBname                  string
	DBport                  string
	HTTPSport               string
	MaxCPUs                 int
	RandomGeneratorsWorkers int
}

func getConfiguration(configurationFilePath string) *Configuration {
	file, error := os.Open(configurationFilePath)
	if error != nil {
		log.Fatal("Configuration file error", error.Error())
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	error = decoder.Decode(&configuration)
	if error != nil {
		log.Fatal("error:", error)
	}
	return &configuration
}
