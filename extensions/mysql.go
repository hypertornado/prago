package extensions

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hypertornado/prago"
)

type Mysql struct{}

func (m *Mysql) Init(app *prago.App) error {
	db, err := ConnectMysql(
		app.Config().GetString("dbUser"),
		app.Config().GetString("dbName"),
		app.Config().GetString("dbPassword"),
	)
	if err != nil {
		return err
	}

	app.Data()["db"] = db
	return nil
}

func ConnectMysql(user, password, dbName string) (*sql.DB, error) {
	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", user, password, dbName)
	db, err := sql.Open("mysql", connectString)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while opening MySQL database: %s", err))
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while ping to MySQL database: %s", err))
	}

	return db, nil
}
