package prago

import (
	"github.com/Sirupsen/logrus"
	"os"
	"time"
)

func createLogger(dotPath string, developmentMode bool) *logrus.Logger {
	logger := logrus.New()
	logFormatter := new(logrus.TextFormatter)
	logFormatter.FullTimestamp = true
	logger.Formatter = logFormatter

	if developmentMode {
		logger.Out = os.Stdout
	} else {
		logPath := dotPath + "/prago.log"
		file, err := os.OpenFile(logPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
		if err != nil {
			panic(err)
		}
		logger.Out = file
	}
	return logger
}

func timestampLog(request Request, text string) {
	if request.Request().Header.Get("X-Dont-Log") != "true" {
		duration := time.Now().Sub(request.receivedAt)
		request.Log().WithField("uuid", request.uuid).WithField("took", duration).
			Println(text)
	}

}
