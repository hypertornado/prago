package extensions

import (
	"database/sql"
	"fmt"
	//use mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/hypertornado/prago"
	"os"
	"os/exec"
)

//Mysql middleware
type Mysql struct{}

//Init Mysql middleware
func (m *Mysql) Init(app *prago.App) error {
	dbUser := app.Config().GetString("dbUser")
	dbPassword := app.Config().GetString("dbPassword")
	dbName := app.Config().GetString("dbName")
	db, err := ConnectMysql(dbUser, dbPassword, dbName)
	if err != nil {
		return err
	}
	app.Data()["db"] = db

	dumpCommand := app.CreateCommand("dump", "Dump database")
	app.AddCommand(dumpCommand, func(app *prago.App) error {
		cmd := exec.Command("mysqldump", "-u"+dbUser, "-p"+dbPassword, dbName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			panic(err)
		}
		return nil
	})

	return nil
}

//ConnectMysql connects to mysql database and returns connection
func ConnectMysql(dbUser, dbPassword, dbName string) (*sql.DB, error) {
	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPassword, dbName)
	db, err := sql.Open("mysql", connectString)
	if err != nil {
		return nil, fmt.Errorf("Error while opening MySQL database: %s", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Error while ping to MySQL database: %s", err)
	}

	return db, nil
}
