package prago

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

type Timeline struct {
	uuid       string
	name       func(string) string
	dataSource func(time.Time, time.Time) float64
	permission Permission
}

type dashboardViewTimeline struct {
	UUID string
	Name string
}

func (app *App) initTimeline() {

	app.API("timeline").Method("GET").Permission(loggedPermission).HandlerJSON(
		func(request *Request) any {
			uuid := request.Param("uuid")
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

func (dashboard *Dashboard) Timeline(name func(string) string, dataSource func(time.Time, time.Time) float64, permission Permission) *Timeline {
	timeline := &Timeline{
		uuid:       "timeline-" + randomString(30),
		name:       name,
		dataSource: dataSource,
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

func (timeline *Timeline) data(request *Request) *timelineData {
	ret := &timelineData{}

	var err error
	var columnsCount int
	columnsCount, err = strconv.Atoi(request.Param("columns"))

	if columnsCount < 10 {
		columnsCount = 10
	}
	if columnsCount > 100 {
		columnsCount = 100
	}

	endDate, err := time.Parse("2006-01-02", request.Param("date"))
	must(err)

	for i := columnsCount - 1; i >= 0; i-- {
		t1 := endDate.AddDate(0, 0, -i)
		t2 := t1.AddDate(0, 0, 1)
		val := timeline.dataSource(t1, t2)

		var isCurrent bool
		if t1.Year() == time.Now().Year() && t1.YearDay() == time.Now().YearDay() {
			isCurrent = true
		}

		ret.Values = append(ret.Values, &timelineDataValue{
			Name:      t1.Format("2. 1. 2006"),
			Value:     val,
			IsCurrent: isCurrent,
		})
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
		if v.Value > maxValue {
			maxValue = v.Value
		}
		if v.Value < td.MinValue {
			td.MinValue = v.Value
		}

		v.ValueText = fmt.Sprintf("%v", v.Value)

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

	}
}

type timelineDataValue struct {
	Name      string
	Value     float64
	ValueText string
	StyleCSS  string
	IsCurrent bool
}
