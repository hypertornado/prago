package prago

import (
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
)

type middlewareLogger struct {
	file   *os.File
	logger *logrus.Logger
}

func (m *middlewareLogger) Init(app *App) error {
	err := os.MkdirAll(app.dotPath+"/log", 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}

	m.logger = logrus.New()
	logFormatter := new(logrus.TextFormatter)
	logFormatter.FullTimestamp = true
	m.logger.Formatter = logFormatter
	app.logger = m.logger

	m.file = m.openLogFile(app)
	m.logger.Out = m.file

	app.AddCronTask("remove old log files", func() {
		m.removeLogFiles(app, time.Now().AddDate(0, 0, -7))
	}, func(in time.Time) time.Time {
		return in.Add(1 * time.Hour)
	})

	app.AddCronTask("rotate log files", func() {
		app.Log().Println("Rotating log files")
		if m.file != nil {
			newFile := m.openLogFile(app)
			oldFile := m.file
			m.logger.Out = newFile
			m.file = newFile
			oldFile.Close()
		}
	}, func(in time.Time) time.Time {
		return in.Add(24 * time.Hour)
	})
	app.requestMiddlewares = append(app.requestMiddlewares, requestMiddlewareLogBefore)
	return nil
}

func (m *middlewareLogger) setStdOut() {
	if m.file != nil {
		m.logger.Out = os.Stdout
		m.file.Close()
		m.file = nil
	}
}

func (m *middlewareLogger) removeLogFiles(app *App, deadline time.Time) {
	logPath := app.dotPath + "/log"
	files, err := ioutil.ReadDir(logPath)
	if err != nil {
		app.Log().Println("error while removing old logs:", err)
		return
	}

	for _, file := range files {
		if file.ModTime().Before(deadline) {
			removePath := logPath + "/" + file.Name()
			err := os.Remove(removePath)
			if err != nil {
				app.Log().Println("Error while removing old log file:", err)
			}
		}
	}
}

func (m *middlewareLogger) openLogFile(app *App) (file *os.File) {
	var err error

	logPath := app.dotPath + "/log/" + time.Now().Format("2006-01-02_15:04:05") + ".log"
	file, err = os.OpenFile(logPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		panic(err)
	}
	return
}

func requestMiddlewareLogBefore(r Request, next func()) {
	r.App().Log().Println(r.Request().Method, r.Request().URL.String())
	next()
}
