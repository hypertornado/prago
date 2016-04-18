package prago

import (
	"github.com/Sirupsen/logrus"
)

type MiddlewareLogger struct{}

func (m MiddlewareLogger) Init(app *App) error {
	logger := logrus.New()
	logFormatter := new(logrus.TextFormatter)
	logFormatter.FullTimestamp = true
	logger.Formatter = logFormatter
	app.logger = logger

	app.requestMiddlewares = append(app.requestMiddlewares, requestMiddlewareLogBefore)

	return nil
}

func requestMiddlewareLogBefore(r Request, next func()) {
	r.App().Log().Println(r.Request().Method, r.Request().URL.String())
	next()
}
