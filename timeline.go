package prago

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

type Timeline struct {
	uuid        string
	name        func(string) string
	permission  Permission
	dataSources []*TimelineDataSource
}

type dashboardViewTimeline struct {
	UUID   string
	Name   string
	Legend *timelineLegend
}

func (app *App) initTimeline() {

	app.API("timeline").Method("GET").Permission(loggedPermission).HandlerJSON(
		func(request *Request) any {
			uuid := request.Param("_uuid")
			timeline, err := app.getTimelineData(request, uuid)
			if err != nil {
				if err == cantFindTimelineError {
					request.WriteJSON(404, "can't find timeline")
					return nil
				}
				panic(err)
			}
			return timeline
		},
	)

}

func (dashboard *Dashboard) Timeline(name func(string) string, permission Permission) *Timeline {
	timeline := &Timeline{
		uuid:       "timeline-" + randomString(30),
		name:       name,
		permission: permission,
	}
	dashboard.board.app.dashboardTimelineMap[timeline.uuid] = timeline
	dashboard.timelines = append(dashboard.timelines, timeline)
	return timeline
}

var cantFindTimelineError = errors.New("can't find timeline")

func (app *App) getTimelineData(request *Request, uuid string) (*timelineData, error) {
	timeline := app.dashboardTimelineMap[uuid]
	if timeline == nil {
		return nil, cantFindTimelineError
	}
	if !request.Authorize(timeline.permission) {
		return nil, errors.New("can't authorize for access of timeline data")
	}
	return timeline.data(request), nil
}

func (timeline *Timeline) getTimelineColumnsCount(widthStr string) int {
	var defaultValue = 10
	var optimalSize = 40
	width, err := strconv.Atoi(widthStr)
	if err != nil {
		return defaultValue
	}
	var barCount = len(timeline.dataSources)
	if barCount == 0 {
		return defaultValue
	}

	return int(math.Floor(
		float64(width) / float64(barCount*optimalSize),
	))

}

func (timeline *Timeline) data(request *Request) *timelineData {
	ret := &timelineData{}

	var columnsCount = timeline.getTimelineColumnsCount(request.Param("_width"))
	var dateStr = request.Param("_date")

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

	for i := columnsCount - 1; i >= 0; i-- {
		var t1, t2 time.Time
		var formattedDate string
		var isCurrent bool

		if typ == "day" {
			t1 = endDate.AddDate(0, 0, -i)
			t2 = t1.AddDate(0, 0, 1)
			formattedDate = t1.Format("2. 1. 2006")
			if t1.Year() == time.Now().Year() && t1.YearDay() == time.Now().YearDay() {
				isCurrent = true
			}
		}
		if typ == "month" {
			t1 = endDate.AddDate(0, -i, 0)
			t2 = t1.AddDate(0, 1, 0)
			formattedDate = t1.Format("1. 2006")
			if t1.Year() == time.Now().Year() && t1.Month() == time.Now().Month() {
				isCurrent = true
			}
		}
		if typ == "year" {
			t1 = endDate.AddDate(-i, 0, 0)
			t2 = t1.AddDate(1, 0, 0)
			formattedDate = t1.Format("2006")
			if t1.Year() == time.Now().Year() {
				isCurrent = true
			}
		}

		dataValue := &timelineDataValue{
			Name:      formattedDate,
			IsCurrent: isCurrent,
		}

		for _, ds := range timeline.dataSources {
			val := ds.dataSource(t1, t2)
			dataValue.Bars = append(dataValue.Bars, &timelineDataBar{
				Value:     val,
				ValueText: ds.stringer(val),
				Color:     ds.color,
			})
		}

		ret.Values = append(ret.Values, dataValue)
	}

	ret.FixValues()
	return ret
}

type timelineData struct {
	Values   []*timelineDataValue
	MinValue float64
	MaxValue float64
}

func (td *timelineData) FixValues() {
	var maxValue float64 = -math.MaxFloat64
	for _, v := range td.Values {
		for _, bar := range v.Bars {
			if bar.Value > maxValue {
				maxValue = bar.Value
			}
			if bar.Value < td.MinValue {
				td.MinValue = bar.Value
			}
			//bar.ValueText = fmt.Sprintf("%v", bar.Value)
		}
	}
	td.MaxValue = maxValue

	for _, v := range td.Values {
		for _, bar := range v.Bars {
			var height, bottom float64
			var size = td.MaxValue - td.MinValue

			if math.Abs(bar.Value) > 0 {
				height = (math.Abs(bar.Value) / size) * 100
			}

			if td.MinValue < 0 {

				var bottomSize = -td.MinValue
				if bar.Value < 0 {
					bottomSize += bar.Value
				}

				bottom = (bottomSize / size) * 100
			}

			bar.StyleCSS = fmt.Sprintf("height: %v%%; bottom: %v%%; background: %s;", height, bottom, bar.Color)

			var labelBottom = bottom
			if bar.Value > 0 {
				labelBottom += height
			}

			bar.LabelStyleCSS = fmt.Sprintf("bottom: %v%%;", labelBottom)
		}

	}
}

type timelineDataValue struct {
	Name      string
	Bars      []*timelineDataBar
	IsCurrent bool
}

type timelineDataBar struct {
	Value         float64
	ValueText     string
	Color         string
	StyleCSS      string
	LabelStyleCSS string
}
