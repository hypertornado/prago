package prago

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" //use mysql
)

func connectMysql(dbUser, dbPassword, dbName string) (*sql.DB, error) {
	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPassword, dbName)
	db, err := sql.Open("mysql", connectString)
	if err != nil {
		return nil, fmt.Errorf("Error while opening MySQL database: %s", err)
	}

	return db, db.Ping()
}