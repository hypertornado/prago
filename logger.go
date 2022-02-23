package prago

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/hypertornado/prago/pragelastic"
)

type logger struct {
	app    *App
	output io.Writer
	index  *pragelastic.Index[logItem]
}

func newLogger(app *App) *logger {
	ret := &logger{
		app:    app,
		output: os.Stdout,
	}
	return ret
}

func (l *logger) writeString(typ, str string) {
	go func() {
		if l.index != nil {
			err := l.index.UpdateSingle(&logItem{
				ID:   randomString(10),
				Time: time.Now(),
				Typ:  typ,
				Text: str,
			})
			if err != nil {
				fmt.Printf("Logger error, can't update: %s", err)
				return
			}
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

type logItem struct {
	ID   string
	Time time.Time
	Typ  string `elastic-datatype:"keyword"`
	Text string `elastic-datatype:"text"`
}

//https://www.elastic.co/guide/en/elasticsearch/reference/current/text.html#match-only-text-field-type

func (app *App) initLogger() {
	if app.ElasticClient == nil {
		return
	}

	index := pragelastic.NewIndex[logItem](app.ElasticClient)

	tg := app.TaskGroup(unlocalized("Logger"))
	tg.Task(unlocalized("reindex log index")).Handler(func(ta *TaskActivity) error {
		index.Delete()
		return index.Create()
	}).Permission("sysadmin")

	app.logger.index = index

	app.FormAction("log_search").Name(unlocalized("Log")).Permission("sysadmin").Form(func(f *Form, r *Request) {
		f.Title = "Logger"
		f.AddTextInput("q", "Query")
		f.AddSelect("typ", "Typ", [][2]string{
			{"", ""},
			{"panic", "panic"},
			{"error", "error"},
			{"info", "info"},
			{"access", "access"},
		})
		//f.AddDateTimePicker("from_date", "Čas od")
		//f.AddDateTimePicker("to_date", "Čas do")
		f.AddTextInput("size", "Results count").Value = "20"
		f.AddTextInput("offset", "Offset").Value = "0"
		f.AddSubmit("Hledat")
	}).Validation(func(vc ValidationContext) {
		query := index.Query().Sort("Time", false).Limit(30)

		fmt.Println(vc.GetValue("from_date"))

		/*from, err := time.Parse("2006-01-02T15:04", vc.GetValue("from_date"))
		if err != nil {
			fmt.Println(from)
		}*/

		size, err := strconv.Atoi(vc.GetValue("size"))
		if err != nil || size <= 0 {
			vc.AddItemError("size", "Must be positive number")
		}

		offset, err := strconv.Atoi(vc.GetValue("offset"))
		if err != nil || offset < 0 {
			vc.AddItemError("offset", "Must be non negative number")
		}

		if vc.GetValue("q") != "" {
			query.Should("Text", vc.GetValue("q"))
		}

		typ := vc.GetValue("typ")
		if typ != "" {
			query.Filter("Typ", typ)
		}

		items, total, err := query.Limit(int64(size)).Offset(int64(offset)).List()
		if err != nil {
			vc.AddError(err.Error())
			return
		}

		table := app.Table()
		table.Header("ID", "Datum", "Typ", "Text")

		table.AddFooterText(fmt.Sprintf("Celkem %d záznamů", total))

		for _, v := range items {
			table.Row(v.ID, v.Time.Format("2. 1. 2006 15:04:05"), v.Typ, v.Text)
		}

		vc.Validation().AfterContent = table.ExecuteHTML()
	})
}