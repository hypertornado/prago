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

	Lines []*timelineDataLine

	Filters []*TimelineDataFilter
}

type timelineDataValue struct {
	DateID     string
	Name       string
	IsCurrent  bool
	IsSelected bool

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
	var size = td.MaxValue - td.MinValue

	for _, v := range td.Values {
		var height, bottom float64

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

	td.Lines = getTimelineDataLines(td.MinValue, td.MaxValue)
}

func (timeline *Timeline) getTimelineData(request *Request, tr *TimelineRequest) *timelineData {
	ret := &timelineData{}
	var columnsCount = timeline.getTimelineColumnsCount(tr.Width)

	language := request.Locale()

	var err error
	var typ string
	var endDate time.Time

	endDate, err = time.Parse("2006-01-02", tr.DateStr)
	if err == nil {
		typ = "day"
	}

	if typ == "" {
		endDate, err = time.Parse("2006-01", tr.DateStr)
		if err == nil {
			typ = "month"
		}
	}

	if typ == "" {
		endDate, err = time.Parse("2006", tr.DateStr)
		if err == nil {
			typ = "year"
		}
	}

	if typ == "" {
		panic("can't parse date")
	}

	var dateIntervals []*timelineDateInterval

	shiftFrom := -columnsCount + 1
	shiftTo := 0

	if tr.Alignment == "future" {
		shiftFrom = 0
		shiftTo = columnsCount - 1
	}
	if tr.Alignment == "center" {
		half := int(columnsCount / 2)
		shiftFrom = -half
		shiftTo = columnsCount - half - 1
	}

	for i := shiftFrom; i <= shiftTo; i++ {
		dateInterval := getTimelineDateInterval(typ, endDate, int64(i))
		dateIntervals = append(dateIntervals, dateInterval)

		dataValue := &timelineDataValue{
			DateID:     dateInterval.DateID,
			Name:       dateInterval.FormattedDate,
			IsCurrent:  dateInterval.IsCurrent,
			IsSelected: dateInterval.IsSelected,
		}
		ret.Values = append(ret.Values, dataValue)
	}

	for k := range dateIntervals {

		var cacheHit bool
		if tr.ValueCache != nil {
			var cacheValue float64
			cacheValue, cacheHit = tr.ValueCache[ret.Values[k].DateID]
			if cacheHit {
				ret.Values[k].Value = cacheValue
			}
		}

		if !cacheHit {
			ret.Values[k].Value = timeline.dataSource(
				&TimelineDataRequest{
					From:    dateIntervals[k].From,
					To:      dateIntervals[k].To,
					Options: tr.Options,
					Request: request,
				},
			)
		}
	}

	for _, value := range ret.Values {
		strVal := humanizeFloat(value.Value, language)
		if timeline.unit != nil {
			strVal += " " + timeline.unit(language)
		}
		value.ValueText = strVal
	}

	timeline.setDataFilter(request, ret, tr.Options)
	ret.fixValues()
	return ret
}

type timelineDateInterval struct {
	DateID        string
	From          time.Time
	To            time.Time
	FormattedDate string
	IsCurrent     bool
	IsSelected    bool
}

func getTimelineDateInterval(typ string, endDate time.Time, shift int64) *timelineDateInterval {
	var t1, t2 time.Time
	var formattedDate string
	var isCurrent bool
	var dateID string

	if typ == "day" {
		t1 = endDate.AddDate(0, 0, int(shift))
		t2 = t1.AddDate(0, 0, 1)
		dateID = t1.Format("2006-01-02")
		formattedDate = t1.Format("2. 1. 2006")
		if t1.Year() == time.Now().Year() && t1.YearDay() == time.Now().YearDay() {
			isCurrent = true
		}
	}
	if typ == "month" {
		t1 = endDate.AddDate(0, int(shift), 0)
		t2 = t1.AddDate(0, 1, 0)
		dateID = t1.Format("2006-01")
		formattedDate = monthName(int64(t1.Month()), "cs") + " " + t1.Format("2006")
		if t1.Year() == time.Now().Year() && t1.Month() == time.Now().Month() {
			isCurrent = true
		}
	}
	if typ == "year" {
		t1 = endDate.AddDate(int(shift), 0, 0)
		t2 = t1.AddDate(1, 0, 0)
		dateID = t1.Format("2006")
		formattedDate = "Rok " + t1.Format("2006")
		if t1.Year() == time.Now().Year() {
			isCurrent = true
		}
	}

	var isSelected bool
	if shift == 0 {
		isSelected = true
	}

	return &timelineDateInterval{
		DateID:        dateID,
		From:          t1,
		To:            t2,
		FormattedDate: formattedDate,
		IsCurrent:     isCurrent,
		IsSelected:    isSelected,
	}
}
