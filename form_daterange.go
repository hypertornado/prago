package prago

import (
	"fmt"
	"strings"
	"time"
)

type dateRange struct {
	Name      func(string) string
	DateRange func() [2]time.Time
	Highlight bool
}

var defaultDateRanges = []dateRange{
	{
		Name: unlocalized("Předevčírem"),
		DateRange: func() [2]time.Time {
			return [2]time.Time{time.Now().AddDate(0, 0, -2), time.Now().AddDate(0, 0, -2)}
		},
	},
	{
		Name: unlocalized("Včera"),
		DateRange: func() [2]time.Time {
			return [2]time.Time{time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, -1)}
		},
	},
	{
		Name: unlocalized("Dnes"),
		DateRange: func() [2]time.Time {
			return [2]time.Time{time.Now(), time.Now()}
		},
		Highlight: true,
	},
	{
		Name: unlocalized("Zítra"),
		DateRange: func() [2]time.Time {
			return [2]time.Time{time.Now().AddDate(0, 0, 1), time.Now().AddDate(0, 0, 1)}
		},
	},
	{
		Name: unlocalized("Pozítří"),
		DateRange: func() [2]time.Time {
			return [2]time.Time{time.Now().AddDate(0, 0, 2), time.Now().AddDate(0, 0, 2)}
		},
	},
	{
		Name:      relativeMonthDateRangeName(-2),
		DateRange: relativeMonthDateRange(-2),
	},
	{
		Name:      relativeMonthDateRangeName(-1),
		DateRange: relativeMonthDateRange(-1),
	},
	{
		Name:      relativeMonthDateRangeName(0),
		DateRange: relativeMonthDateRange(0),
		Highlight: true,
	},
	{
		Name:      relativeMonthDateRangeName(1),
		DateRange: relativeMonthDateRange(1),
	},
	{
		Name:      relativeMonthDateRangeName(2),
		DateRange: relativeMonthDateRange(2),
	},
	{
		Name:      relativeYearDateRangeName(-1),
		DateRange: relativeYearDateRange(-1),
	},
	{
		Name:      relativeYearDateRangeName(0),
		DateRange: relativeYearDateRange(0),
		Highlight: true,
	},
	{
		Name:      relativeYearDateRangeName(1),
		DateRange: relativeYearDateRange(1),
	},
}

func relativeMonthDateRangeName(rel int64) func(string) string {
	t := time.Now().AddDate(0, int(rel), 0)
	return func(string) string {
		return fmt.Sprintf("%s %d", monthsCS[t.Month()-1], t.Year())
	}
}

func relativeMonthDateRange(rel int64) func() [2]time.Time {
	t := time.Now().AddDate(0, int(rel), 0)
	return func() [2]time.Time {
		first := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
		last := first.AddDate(0, 1, -1)
		return [2]time.Time{first, last}
	}
}

func relativeYearDateRangeName(rel int64) func(string) string {
	t := time.Now().AddDate(int(rel), 0, 0)
	return func(string) string {
		return fmt.Sprintf("Rok %d", t.Year())
	}
}

func relativeYearDateRange(rel int64) func() [2]time.Time {
	t := time.Now().AddDate(int(rel), 0, 0)
	return func() [2]time.Time {
		first := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
		last := time.Date(t.Year(), 12, 31, 0, 0, 0, 0, t.Location())
		return [2]time.Time{first, last}
	}
}

func (app *App) initDateRange() {

	PopupForm(app, "_dateranges", func(form *Form, request *Request) {
		var options []*FormOption
		for _, dateRange := range defaultDateRanges {
			dates := dateRange.DateRange()
			option := &FormOption{
				ID:   fmt.Sprintf("%s_%s", dates[0].Format("2006-01-02"), dates[1].Format("2006-01-02")),
				Name: dateRange.Name(request.Locale()),
			}
			if dateRange.Highlight {
				option.Style = "create"
				option.Icon = "glyphicons-basic-306-square-empty-play.svg"
			}

			if dates[0].Year() == dates[1].Year() && dates[0].YearDay() == dates[1].YearDay() {
				option.DescriptionAfter = dates[0].Format("2. 1. 2006")
			} else {
				option.DescriptionAfter = fmt.Sprintf("%s – %s", dates[0].Format("2. 1. 2006"), dates[1].Format("2. 1. 2006"))
			}

			options = append(options, option)
		}

		form.AddRadioOptions("date", "", options)
		form.AutosubmitOnDataChange = true

	}, func(fv FormValidation, request *Request) {

		dateStr := request.Param("date")

		if dateStr == "" {
			return
		}
		fv.Data(dateStr)
	}).Permission("everybody").Name(unlocalized("Vybrat interval")).Icon("glyphicons-basic-58-history.svg")

}

func (form *Form) AddDateRange(name, description string, from, to time.Time) *FormItem {
	input := form.addInput(name, description, "form_input_daterange")
	input.Value = fmt.Sprintf("%s_%s", from.Format("2006-01-02"), to.Format("2006-01-02"))
	return input
}

func (fi *FormItem) DateRangeFrom() string {
	items := strings.Split(fi.Value, "_")
	if len(items) != 2 {
		return ""
	}
	return items[0]
}

func (fi *FormItem) DateRangeTo() string {
	items := strings.Split(fi.Value, "_")
	if len(items) != 2 {
		return ""
	}
	return items[1]
}

func ParseDateRange(request *Request, itemID string) *[2]time.Time {
	fromStr := request.Param(itemID + "_from")
	toStr := request.Param(itemID + "_to")

	var ret [2]time.Time
	var err error
	ret[0], err = time.Parse("2006-01-02", fromStr)
	if err != nil {
		return nil
	}
	ret[1], err = time.Parse("2006-01-02", toStr)
	if err != nil {
		return nil
	}

	if ret[0].IsZero() || ret[1].IsZero() {
		return nil
	}

	if ret[1].Before(ret[0]) {
		return nil
	}

	return &ret

}
