package prago

import (
	"context"
	"fmt"
	"time"
)

type TimelineDataSource struct {
	name       func(string) string
	dataSource func(*TimelineDataRequest) float64
	stringer   func(float64) string
	color      string
}

type TimelineDataRequest struct {
	From    time.Time
	To      time.Time
	Context context.Context
}

func (timeline *Timeline) DataSource(dataSource func(request *TimelineDataRequest) float64) *TimelineDataSource {
	ds := &TimelineDataSource{
		name:       timeline.name,
		dataSource: dataSource,
		stringer: func(f float64) string {
			return fmt.Sprintf("%v", f)
		},
		color: getTimelineColor(len(timeline.dataSources)),
	}
	timeline.dataSources = append(timeline.dataSources, ds)
	return ds
}

func (tds *TimelineDataSource) Name(name func(string) string) *TimelineDataSource {
	tds.name = name
	return tds
}

func (tds *TimelineDataSource) Stringer(stringer func(float64) string) *TimelineDataSource {
	tds.stringer = stringer
	return tds
}

func getTimelineColor(order int) string {
	hues := []int{
		214, 334, 94, 274, 34, 154,
	}
	order = order % len(hues)
	return fmt.Sprintf("hsl(%d, 50%%, 50%%)", hues[order])
}
