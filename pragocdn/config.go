package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type CDNConfig struct {
	Profile  string
	Accounts []CDNConfigAccount
}

type CDNConfigAccount struct {
	Name     string
	Password string
}

func loadCDNConfig() (CDNConfig, error) {
	path := os.Getenv("HOME") + "/.pragocdn/cdn-config.json"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return CDNConfig{}, fmt.Errorf("error while opening file %s: %s", path, err)
	}

	var config CDNConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return CDNConfig{}, err
	}

	return config, nil
}
