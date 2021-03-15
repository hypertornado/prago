package prago

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

func (app *App) initConfig() {
	path := getConfigPath(app.codeName)

	kv := make(map[string]interface{})
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("error while opening config file %s: %s", path, err))
	}

	err = json.Unmarshal(data, &kv)
	if err != nil {
		panic(fmt.Sprintf("error while parsing config file: %s", err))
	}
	app.config = config{kv}
}

func getConfigPath(appName string) string {
	return fmt.Sprintf("%s/.%s/config.json", os.Getenv("HOME"), appName)
}

func configExists(appName string) bool {
	_, err := os.Open(getConfigPath(appName))
	if err == nil {
		return true
	}
	return false
}

type config struct {
	v map[string]interface{}
}

//Export outputs config data in human readable form
func (c config) Export() [][2]string {
	keys := []string{}
	for k := range c.v {
		keys = append(keys, k)
	}
	keySlice := sort.StringSlice(keys)
	keySlice.Sort()

	ret := [][2]string{}
	for _, v := range keys {
		ret = append(ret, [2]string{v, fmt.Sprintf("%s", c.v[v])})
	}

	return ret
}

//Get returns config item
func (app *App) ConfigurationGetItem(name string) (interface{}, error) {
	c := app.config
	val, ok := c.v[name]
	if ok {
		return val, nil
	}
	return nil, errors.New("item in config not found")
}

//GetString returns config string item
//panics when item is not set or not a string
func (app *App) ConfigurationGetString(name string) string {
	item, err := app.ConfigurationGetItem(name)
	if err != nil {
		panic(fmt.Sprintf("error while getting '%s': %s", name, err.Error()))
	}
	str, ok := item.(string)
	if !ok {
		panic("config item is not string")
	}
	return str
}

//GetStringWithFallback returns config string with default fallback value
func (app *App) ConfigurationGetStringWithFallback(name, fallback string) string {
	item, err := app.ConfigurationGetItem(name)
	if err != nil {
		return fallback
	}
	str, ok := item.(string)
	if !ok {
		return fallback
	}
	return str
}
