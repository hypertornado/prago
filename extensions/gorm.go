package extensions

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hypertornado/prago"
	"github.com/jinzhu/gorm"
	"os"
	"os/exec"
)

type Gorm struct {
	DB gorm.DB
}

func (g *Gorm) Init(app *prago.App) error {
	config, err := app.Config()
	if err != nil {
		return err
	}

	user := config["dbUser"]
	dbName := config["dbName"]
	password := config["dbPassword"]
	db, err := g.connectMySQL(user, password, dbName)
	app.Data()["db"] = db.DB()
	app.Data()["gorm"] = &db
	g.DB = db

	dumpCommand := app.CreateCommand("dump", "Dump database")
	app.AddCommand(dumpCommand, func(app *prago.App) error {

		cmd := exec.Command("mysqldump", "-u"+user, "-p"+password, dbName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			panic(err)
		}
		return nil
	})

	return err
}

func (g Gorm) connectMySQL(user, password, dbName string) (gorm.DB, error) {
	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", user, password, dbName)
	return gorm.Open("mysql", connectString)
}
