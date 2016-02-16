package extensions

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hypertornado/prago"
	"github.com/jinzhu/gorm"
)

type Gorm struct {
	DB gorm.DB
}

func (g *Gorm) Init(app *prago.App) error {
	config, err := app.Config()
	if err != nil {
		return err
	}

	user := app.Data()["appName"].(string)
	dbName := app.Data()["appName"].(string)
	password := config["dbPassword"]
	db, err := g.connectMySQL(user, password, dbName)
	app.Data()["db"] = db.DB()
	app.Data()["gorm"] = db
	g.DB = db
	return err
}

func (g Gorm) connectMySQL(user, password, dbName string) (gorm.DB, error) {
	connectString := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", user, password, dbName)
	return gorm.Open("mysql", connectString)
}
