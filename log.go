package prago

import (
	"github.com/Sirupsen/logrus"
	"os"
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
