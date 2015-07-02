package prago

import (
	"github.com/Machiel/slugify"
	"github.com/Sirupsen/logrus"
	"os"
)

func defaultLogger() *logrus.Logger {
	ret := logrus.New()
	logFormatter := new(logrus.TextFormatter)
	logFormatter.FullTimestamp = true
	ret.Formatter = logFormatter
	return ret
}

func writeStartInfo(log *logrus.Logger, port int, developmentMode bool) {
	log.WithField("port", port).
		WithField("pid", os.Getpid()).
		WithField("development mode", developmentMode).
		Info("Server started")
}

func Redirect(request Request, urlStr string) {
	request.Header().Set("Location", urlStr)
}

func PrettyUrl(s string) string {
	return slugify.Slugify(s)
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}
