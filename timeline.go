package prago

import (
	"errors"
	"math"
	"strconv"
	"time"

	"golang.org/x/net/context"
)

type Timeline struct {
	uuid       string
	name       func(string) string
	permission Permission

	dataSource func(*TimelineDataRequest) float64

	unit func(string) string
}

type dashboardViewTimeline struct {
	UUID string
	Name string
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

func (dashboard *Dashboard) Timeline(name func(string) string, permission Permission, dataSource func(request *TimelineDataRequest) float64) *Timeline {
	timeline := &Timeline{
		uuid:       "timeline-" + randomString(30),
		name:       name,
		permission: permission,
		dataSource: dataSource,
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
	return timeline.getTimelineData(request), nil
}

func (timeline *Timeline) getTimelineColumnsCount(widthStr string) int {
	var optimalSize = 20
	var paddingLeft = 60
	width, err := strconv.Atoi(widthStr)
	must(err)
	width = width - paddingLeft
	return int(math.Floor(
		float64(width) / float64(optimalSize),
	))

}

type TimelineDataRequest struct {
	From    time.Time
	To      time.Time
	Context context.Context
}

func (timeline *Timeline) Unit(unit func(string) string) *Timeline {
	timeline.unit = unit
	return timeline
}
