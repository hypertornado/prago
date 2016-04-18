package prago

import (
	//"fmt"
	"github.com/Sirupsen/logrus"
	"os"
)

type MiddlewareLogger struct{}

func (m MiddlewareLogger) Init(app *App) error {
	var err error

	err = os.Mkdir(app.dotPath+"/log", 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}

	logPath := app.dotPath + "/log/default.log"

	var file *os.File

	file, err = os.OpenFile(logPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		if os.IsExist(err) {
			file, err = os.Create(logPath)
		}
	}

	logger := logrus.New()
	logFormatter := new(logrus.TextFormatter)
	logFormatter.FullTimestamp = true
	logger.Formatter = logFormatter
	app.logger = logger

	logger.Out = file

	app.requestMiddlewares = append(app.requestMiddlewares, requestMiddlewareLogBefore)

	return nil
}

func requestMiddlewareLogBefore(r Request, next func()) {
	r.App().Log().Println(r.Request().Method, r.Request().URL.String())
	next()
}
