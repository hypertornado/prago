package prago

import (
	"fmt"
	"io"
	"os"
	"time"
)

type logger struct {
	output io.Writer
}

func newLogger() *logger {
	ret := &logger{
		output: os.Stdout,
	}
	return ret
}

func (l *logger) writeString(str string) {
	str = fmt.Sprintf("%s %s\n", time.Now().Format("2006/01/02 15:04:05"), str)
	_, err := l.output.Write([]byte(str))
	if err != nil {
		panic(err)
	}
}

func (l *logger) Println(v ...any) {
	l.writeString(fmt.Sprintln(v...))
}

func (l *logger) Printf(format string, a ...any) {
	l.writeString(fmt.Sprintf(format, a...))
}

func (l *logger) SetOutput(w io.Writer) {
	l.output = w
}

func (l *logger) Fatal(v ...any) {
	l.Println(v...)
	os.Exit(1)
}

func (l *logger) Fatalf(format string, v ...any) {
	l.Printf(format, v...)
	os.Exit(1)
}
