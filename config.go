package prago

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

func loadConfig(appName string) config {
	path := fmt.Sprintf("%s/.%s/config.json", os.Getenv("HOME"), appName)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("error while opening config file %s: %s", path, err))
	}

	kv := make(map[string]interface{})

	err = json.Unmarshal(data, &kv)
	if err != nil {
		panic(fmt.Sprintf("error while parsing config file: %s", err))
	}

	return config{kv}
}

type config struct {
	v map[string]interface{}
}

func (c *config) Set(k string, val interface{}) {
	c.v[k] = val
}

//Export outputs config data in human readable form
func (c *config) Export() [][2]string {
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
func (c *config) Get(name string) (interface{}, error) {
	val, ok := c.v[name]
	if ok {
		return val, nil
	}
	return nil, errors.New("item in config not found")
}

//GetString returns config string item
//panics when item is not set or not a string
func (c *config) GetString(name string) string {
	item, err := c.Get(name)
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
func (c *config) GetStringWithFallback(name, fallback string) string {
	item, err := c.Get(name)
	if err != nil {
		return fallback
	}
	str, ok := item.(string)
	if !ok {
		return fallback
	}
	return str
}
