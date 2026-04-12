package prago

import (
	"errors"
	"math"
	"time"

	"golang.org/x/net/context"
)

type Timeline struct {
	uuid       string
	name       func(string) string
	permission Permission

	dataSource func(*TimelineDataRequest) float64

	unit func(string) string

	defaultAlignment string

	optionsForm func(request *Request, form *Form)

	filterCustomNames map[string]func(string) (string, string)
}

type dashboardViewTimeline struct {
	UUID      string
	Name      string
	Alignment string
}

type TimelineRequest struct {
	UUID       string
	DateStr    string
	Width      int64
	ValueCache map[string]float64
	Alignment  string
	Options    map[string]string
}

func (app *App) initTimeline() {

	APIJSON(app, "timeline", func(request *Request, tr *TimelineRequest) any {
		timeline, err := app.getTimelineData(request, tr)
		if err != nil {
			if err == cantFindTimelineError {
				request.WriteJSON(404, "can't find timeline")
				return nil
			}
			panic(err)
		}
		return timeline
	}).Permission(loggedPermission).Method("POST")

	app.initTimelineSettings()
}

func (dashboard *Dashboard) Timeline(name func(string) string, permission Permission, dataSource func(request *TimelineDataRequest) float64) *Timeline {
	timeline := &Timeline{
		uuid:              "timeline-" + randomString(30),
		name:              name,
		permission:        permission,
		dataSource:        dataSource,
		defaultAlignment:  "history",
		filterCustomNames: map[string]func(string) (string, string){},
	}
	dashboard.board.app.dashboardTimelineMap[timeline.uuid] = timeline
	dashboard.timelines = append(dashboard.timelines, timeline)
	return timeline
}

var cantFindTimelineError = errors.New("can't find timeline")

func (app *App) getTimeline(request *Request, uuid string) (*Timeline, error) {
	timeline := app.dashboardTimelineMap[uuid]
	if timeline == nil {
		return nil, cantFindTimelineError
	}
	if !request.Authorize(timeline.permission) {
		return nil, errors.New("can't authorize for access of timeline data")
	}
	return timeline, nil
}

func (app *App) getTimelineData(request *Request, tr *TimelineRequest) (*timelineData, error) {
	timeline, err := app.getTimeline(request, tr.UUID)
	if err != nil {
		return nil, err
	}
	return timeline.getTimelineData(request, tr), nil
}

func (timeline *Timeline) getTimelineColumnsCount(width int64) int {
	var optimalSize = 20
	var paddingLeft = 60
	width = width - int64(paddingLeft)
	return int(math.Floor(
		float64(width) / float64(optimalSize),
	))

}

type TimelineDataRequest struct {
	From    time.Time
	To      time.Time
	Context context.Context
	Options map[string]string
}

func (timeline *Timeline) Unit(unit func(string) string) *Timeline {
	timeline.unit = unit
	return timeline
}

func (timeline *Timeline) Future() *Timeline {
	timeline.defaultAlignment = "future"
	return timeline
}

func (timeline *Timeline) Center() *Timeline {
	timeline.defaultAlignment = "center"
	return timeline
}

func (timeline *Timeline) OptionsForm(fn func(request *Request, form *Form)) *Timeline {
	timeline.optionsForm = fn
	return timeline
}

func (timeline *Timeline) FilterName(key string, fn func(value string) (string, string)) *Timeline {
	timeline.filterCustomNames[key] = fn
	return timeline
}
