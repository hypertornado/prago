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

func (app *App) connectDB() {
	path := getDBConnectPath(app.codeName)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("error while opening db config file %s: %s", path, err))
	}

	var config dbConnectConfig

	err = json.Unmarshal(data, &config)
	if err != nil {
		panic(fmt.Sprintf("error while parsing db config file: %s", err))
	}

	app.dbConfig = &config

	app.db = mustConnectDatabase(
		app.dbConfig.User,
		app.dbConfig.Password,
		app.dbConfig.Name,
	)

}
