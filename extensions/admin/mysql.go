package admin

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hypertornado/prago"
)

func connectMysql(app *prago.App) *sql.DB {
	dbUser := app.Config.GetStringWithFallback("dbUser", "")
	dbPassword := app.Config.GetStringWithFallback("dbPassword", "")
	dbName := app.Config.GetStringWithFallback("dbName", "")

	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPassword, dbName)
	db, err := sql.Open("mysql", connectString)
	if err != nil {
		app.Log().Fatalf("Error while opening MySQL database: %s", err)
	}
	return db
}
