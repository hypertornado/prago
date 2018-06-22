package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	Servers []ConfigServer
}

type ConfigServer struct {
	Name string
	Host string
}

func loadConfig() (Config, error) {
	path := os.Getenv("HOME") + "/.pragoserver"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("error while opening file %s: %s", path, err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
