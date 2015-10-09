package prago

import (
	"github.com/Machiel/slugify"
	"github.com/Sirupsen/logrus"
	"math/rand"
	"os"
	"time"
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

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var seeded = false

func RandomString(n int) string {
	if !seeded {
		rand.Seed(time.Now().Unix())
		seeded = true
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
