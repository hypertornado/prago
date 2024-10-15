package prago

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" //use mysql
)

func mustConnectDatabase(dbUser, dbPassword, dbName string) *sql.DB {
	db, err := connectMysql(dbUser, dbPassword, dbName)
	if err != nil {
		panic("can't connect to database: " + err.Error())
	}
	return db
}

func connectMysql(dbUser, dbPassword, dbName string) (*sql.DB, error) {
	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbUser, dbPassword, dbName)
	db, err := sql.Open("mysql", connectString)
	if err != nil {
		return nil, fmt.Errorf("error while opening MySQL database: %s", err)
	}
	//prevent resource exhaustion
	//https://github.com/go-sql-driver/mysql#usage
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)

	db.SetConnMaxLifetime(time.Minute * 1)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return db, db.Ping()
}
