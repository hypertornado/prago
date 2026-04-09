package prago

import (
	"fmt"
	"math"
	"time"
)

type timelineData struct {
	Values   []*timelineDataValue
	MinValue float64
	MaxValue float64
}

type timelineDataValue struct {
	Name      string
	IsCurrent bool

	Value         float64
	ValueText     string
	StyleCSS      string
	LabelStyleCSS string
}

func (td *timelineData) fixValues() {
	var maxValue float64 = -math.MaxFloat64
	for _, v := range td.Values {
		if v.Value > maxValue {
			maxValue = v.Value
		}
		if v.Value < td.MinValue {
			td.MinValue = v.Value
		}
	}
	td.MaxValue = maxValue

	for _, v := range td.Values {
		var height, bottom float64
		var size = td.MaxValue - td.MinValue

		if math.Abs(v.Value) > 0 {
			height = (math.Abs(v.Value) / size) * 100
		}

		if td.MinValue < 0 {

			var bottomSize = -td.MinValue
			if v.Value < 0 {
				bottomSize += v.Value
			}

			bottom = (bottomSize / size) * 100
		}

		v.StyleCSS = fmt.Sprintf("height: %v%%; bottom: %v%%;", height, bottom)

		var labelBottom = bottom
		if v.Value > 0 {
			labelBottom += height
		}

		v.LabelStyleCSS = fmt.Sprintf("bottom: %v%%;", labelBottom)
	}
}

func (timeline *Timeline) getTimelineData(request *Request) *timelineData {
	ret := &timelineData{}
	var columnsCount = timeline.getTimelineColumnsCount(request.Param("_width"))
	var dateStr = request.Param("_date")

	language := request.Locale()

	var err error
	var typ string
	var endDate time.Time

	endDate, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		typ = "day"
	}

	if typ == "" {
		endDate, err = time.Parse("2006-01", dateStr)
		if err == nil {
			typ = "month"
		}
	}

	if typ == "" {
		endDate, err = time.Parse("2006", dateStr)
		if err == nil {
			typ = "year"
		}
	}

	if typ == "" {
		panic("can't parse date")
	}

	var dateIntervals []*timelineDateInterval

	for i := columnsCount - 1; i >= 0; i-- {
		dateInterval := getTimelineDateInterval(typ, endDate, int64(i))
		dateIntervals = append(dateIntervals, dateInterval)

		dataValue := &timelineDataValue{
			Name:      dateInterval.FormattedDate,
			IsCurrent: dateInterval.IsCurrent,
		}
		ret.Values = append(ret.Values, dataValue)
	}

	for k := range dateIntervals {
		ret.Values[k].Value = timeline.dataSource(
			&TimelineDataRequest{
				From:    dateIntervals[k].From,
				To:      dateIntervals[k].To,
				Context: request.r.Context(),
			},
		)
	}

	for _, value := range ret.Values {
		strVal := humanizeFloat(value.Value, language)
		if timeline.unit != nil {
			strVal += " " + timeline.unit(language)
		}
		value.ValueText = strVal
	}

	ret.fixValues()
	return ret
}

type timelineDateInterval struct {
	From          time.Time
	To            time.Time
	FormattedDate string
	IsCurrent     bool
}

func getTimelineDateInterval(typ string, endDate time.Time, i int64) *timelineDateInterval {
	var t1, t2 time.Time
	var formattedDate string
	var isCurrent bool

	if typ == "day" {
		t1 = endDate.AddDate(0, 0, int(-i))
		t2 = t1.AddDate(0, 0, 1)
		formattedDate = t1.Format("2. 1. 2006")
		if t1.Year() == time.Now().Year() && t1.YearDay() == time.Now().YearDay() {
			isCurrent = true
		}
	}
	if typ == "month" {
		t1 = endDate.AddDate(0, int(-i), 0)
		t2 = t1.AddDate(0, 1, 0)
		formattedDate = t1.Format("1. 2006")
		if t1.Year() == time.Now().Year() && t1.Month() == time.Now().Month() {
			isCurrent = true
		}
	}
	if typ == "year" {
		t1 = endDate.AddDate(int(-i), 0, 0)
		t2 = t1.AddDate(1, 0, 0)
		formattedDate = t1.Format("2006")
		if t1.Year() == time.Now().Year() {
			isCurrent = true
		}
	}

	return &timelineDateInterval{
		From:          t1,
		To:            t2,
		FormattedDate: formattedDate,
		IsCurrent:     isCurrent,
	}
}
