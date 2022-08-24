package prago

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type dbConnectConfig struct {
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func getDBConnectPath(appName string) string {
	return fmt.Sprintf("%s/.%s/prago_db.json", os.Getenv("HOME"), appName)
}

func getDBConfig(codeName string) (*dbConnectConfig, error) {
	path := getDBConnectPath(codeName)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error while opening db config file %s: %s", path, err)
	}

	var config dbConnectConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error while parsing db config file: %s", err)
	}

	return &config, nil

}

func (app *App) connectDB(testing bool) {
	var config *dbConnectConfig

	if testing {
		config = &dbConnectConfig{
			Name:     "prago_test",
			User:     "prago_test",
			Password: "prago_test",
		}
	} else {
		var err error
		config, err = getDBConfig(app.codeName)
		if err != nil {
			panic(fmt.Sprintf("can't get config file: %s", err.Error()))
		}
	}

	app.dbConfig = config
	app.db = mustConnectDatabase(
		app.dbConfig.User,
		app.dbConfig.Password,
		app.dbConfig.Name,
	)

}
