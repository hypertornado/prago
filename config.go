package prago

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

type MiddlewareConfig struct{}

func (m MiddlewareConfig) Init(app *App) error {
	path := os.Getenv("HOME") + "/." + app.data["appName"].(string) + "/config.json"
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	kv := make(map[string]string)

	err = json.Unmarshal(data, &kv)
	if err != nil {
		return err
	}
	app.data["config"] = kv
	return nil
}

func (a *App) Config() (ret map[string]string, err error) {
	var ok bool
	ret, ok = a.data["config"].(map[string]string)
	if !ok {
		err = errors.New("cant get config")
	}
	return
}