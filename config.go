package prago

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type MiddlewareConfig struct{}

func (m MiddlewareConfig) Init(app *App) error {
	path := os.Getenv("HOME") + "/." + app.data["appName"].(string) + "/config.json"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error while opening file %s: %s", path, err))
	}

	kv := make(map[string]interface{})

	err = json.Unmarshal(data, &kv)
	if err != nil {
		return err
	}
	app.data["config"] = kv

	configCommand := app.CreateCommand("config", "Print app configuration")
	app.AddCommand(configCommand, func(app *App) error {
		for k, v := range kv {
			fmt.Println(k, ":", v)
		}
		return nil
	})

	return nil
}

type Config struct {
	v map[string]interface{}
}

func (c *Config) Get(name string) (interface{}, error) {
	val, ok := c.v[name]
	if ok {
		return val, nil
	}
	return nil, errors.New("Item in config not found")
}

func (c *Config) GetString(name string) string {
	item, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	str, ok := item.(string)
	if !ok {
		panic("Config item is not string")
	}
	return str
}

func (a *App) Config() *Config {
	ret, ok := a.data["config"].(map[string]interface{})
	if !ok {
		panic("cant get config")
	}
	return &Config{ret}
}
