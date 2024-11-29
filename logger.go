package prago

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type logger struct {
	app    *App
	output io.Writer
}

func newLogger(app *App) *logger {
	ret := &logger{
		app:    app,
		output: os.Stdout,
	}

	return ret
}

func (app *App) SetLogHandler(fn func(string, string)) {
	app.logHandler = fn
}

func (l *logger) writeString(typ, str string) {
	go func() {
		str = strings.Trim(str, " \t\r\n")

		if l.app.logHandler != nil {
			l.app.logHandler(typ, str)
			if !l.app.developmentMode {
				return
			}
		}

		str = fmt.Sprintf("%s %s %s\n", time.Now().Format("2006/01/02 15:04:05"), typ, str)
		_, err := l.output.Write([]byte(str))
		if err != nil {
			panic(err)
		}
	}()
}

func (l *logger) accessln(v ...any) {
	l.writeString("access", fmt.Sprintln(v...))
}

func (l *logger) panicln(v ...any) {
	l.writeString("panic", fmt.Sprintln(v...))
}

func (l *logger) Println(v ...any) {
	l.writeString("info", fmt.Sprintln(v...))
}

func (l *logger) Printf(format string, a ...any) {
	l.writeString("info", fmt.Sprintf(format, a...))
}

func (l *logger) Errorln(v ...any) {
	l.writeString("error", fmt.Sprintln(v...))
}

func (l *logger) Errorf(format string, a ...any) {
	l.writeString("error", fmt.Sprintf(format, a...))
}

func (l *logger) SetOutput(w io.Writer) {
	l.output = w
}
