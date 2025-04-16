package prago

import (
	"errors"
	"fmt"
)

type dashboardFigure struct {
	uuid               string
	permission         Permission
	url                string
	valueStr           func(*Request) string
	value              func(*Request) int64
	compareValue       func(*Request) int64
	compareDescription func(string) string
	unit               func(string) string
	name               func(string) string
	refreshTimeSeconds int64
}

type dashboardFigureData struct {
	Value       string
	Description string
	IsGreen     bool
	IsRed       bool
}

func (figure dashboardFigure) data(request *Request) *dashboardFigureData {

	if figure.valueStr != nil {

		var suffix string
		if figure.unit != nil {
			suffix = " " + figure.unit(request.Locale())
		}

		return &dashboardFigureData{
			Value: figure.valueStr(request) + suffix,
			//Description: figure.getDescriptionStr(request, values),
		}
	}

	values := figure.getValues(request)
	ret := &dashboardFigureData{
		Value:       figure.getValueStr(request, values),
		Description: figure.getDescriptionStr(request, values),
	}
	ret.IsGreen, ret.IsRed = figure.getColors(values)
	return ret
}

var cantFindFigureError = errors.New("can't find figure")

func (app *App) getDashboardFigureData(request *Request, uuid string) (*dashboardFigureData, error) {
	figure := app.dashboardFigureMap[uuid]
	if figure == nil {
		return nil, cantFindFigureError
	}
	if !request.Authorize(figure.permission) {
		return nil, errors.New("can't authorize for access of figure data")
	}
	return figure.data(request), nil
}

func (dashboard *Dashboard) Figure(name func(string) string, permission Permission) *dashboardFigure {
	must(dashboard.board.app.validatePermission(permission))
	figure := &dashboardFigure{
		uuid:               randomString(30),
		name:               name,
		permission:         permission,
		refreshTimeSeconds: 60,
	}
	dashboard.board.app.dashboardFigureMap[figure.uuid] = figure
	dashboard.figures = append(dashboard.figures, figure)
	return figure
}

func (figure *dashboardFigure) RefreshTime(seconds int64) *dashboardFigure {
	if seconds < 1 {
		seconds = 1
	}
	figure.refreshTimeSeconds = seconds
	return figure
}

func (item *dashboardFigure) getValues(request *Request) [2]int64 {
	var val, cmpVal int64 = -1, -1
	if item.value != nil {
		val = item.value(request)
	}
	if item.compareValue != nil {
		cmpVal = item.compareValue(request)
	}
	return [2]int64{
		val, cmpVal,
	}

}

func (item *dashboardFigure) Value(value func(*Request) int64) *dashboardFigure {
	item.value = value
	return item
}

func (item *dashboardFigure) ValueString(value func(*Request) string) *dashboardFigure {
	item.valueStr = value
	return item
}

func (item *dashboardFigure) Compare(value func(*Request) int64, description func(string) string) *dashboardFigure {
	item.compareValue = value
	item.compareDescription = description
	return item
}

func (item *dashboardFigure) getValueStr(request *Request, values [2]int64) string {
	ret := "–"
	if item.value != nil {
		val := values[0]
		ret = humanizeNumber(val)
		if item.unit != nil && item.unit(request.Locale()) != "" {
			ret += " " + item.unit(request.Locale())
		}
	}

	return ret
}

func (item *dashboardFigure) getDescriptionStr(request *Request, values [2]int64) string {
	if item.value == nil {
		return ""
	}
	if item.compareValue == nil {
		return ""
	}

	val, compareValue := values[0], values[1]

	diff := val - compareValue
	var ret string
	if diff >= 0 {
		ret = fmt.Sprintf("+%s", humanizeNumber(diff))
	} else {
		ret = humanizeNumber(diff)
	}

	if item.unit(request.Locale()) != "" {
		ret += " " + item.unit(request.Locale())
	}

	if compareValue > 0 {
		percent := fmt.Sprintf("%.2f%%", (100*float64(diff))/float64(compareValue))
		ret += fmt.Sprintf(" (%s)", percent)
	}

	if item.compareDescription(request.Locale()) != "" {
		ret += " " + item.compareDescription(request.Locale())
	}

	return ret
}

func (item *dashboardFigure) view(request *Request) *dashboardViewFigure {
	ret := &dashboardViewFigure{
		UUID:               item.uuid,
		URL:                item.url,
		Name:               item.name(request.Locale()),
		RefreshTimeSeconds: item.refreshTimeSeconds,
	}
	return ret
}

func (item *dashboardFigure) getColors(values [2]int64) (isGreen, isRed bool) {
	val, compareValue := values[0], values[1]
	if val > compareValue {
		isGreen = true
	}
	if val < compareValue {
		isRed = true
	}
	return
}

func (item *dashboardFigure) URL(url string) *dashboardFigure {
	item.url = url
	return item
}

func (item *dashboardFigure) Unit(unit func(string) string) *dashboardFigure {
	item.unit = unit
	return item
}
