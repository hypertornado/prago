package extensions

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hypertornado/prago"
)

type Mysql struct{}

func (m *Mysql) Init(app *prago.App) error {
	user := app.Config().GetString("dbUser")
	dbName := app.Config().GetString("dbName")
	password := app.Config().GetString("dbPassword")

	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", user, password, dbName)

	db, err := sql.Open("mysql", connectString)
	if err != nil {
		return err
	}

	app.Data()["db"] = db

	return nil
}
