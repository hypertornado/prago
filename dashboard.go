package prago

import (
	"errors"
	"fmt"
)

type Dashboard struct {
	board   *Board
	name    func(string) string
	figures []*DashboardFigure
	tables  []*dashboardTable
}

type DashboardFigure struct {
	uuid               string
	permission         Permission
	url                string
	value              func() int64
	compareValue       func() int64
	compareDescription string
	unit               string
	name               string
}

type dashboardTable struct {
	uuid       string
	table      func() *Table
	permission Permission
}

type DashboardFigureData struct {
	Value       string
	Description string
	IsGreen     bool
	IsRed       bool
}

func (figure DashboardFigure) data(app *App) *DashboardFigureData {
	values := figure.getValues(app)
	ret := &DashboardFigureData{
		Value:       figure.getValueStr(values, app),
		Description: figure.getDescriptionStr(values, app),
	}
	ret.IsGreen, ret.IsRed = figure.getColors(values, app)
	return ret
}

func (app *App) initDashboard() {

	app.API("dashboard-table").Method("GET").Permission(loggedPermission).Handler(
		func(request *Request) {
			uuid := request.Param("uuid")
			table, err := app.getDashboardTableData(uuid, request.user)
			must(err)
			request.app.templates.templates.ExecuteTemplate(request.Response(), "admin_form_table", table.TemplateData())
		},
	)

	app.API("dashboard-figure").Method("GET").Permission(loggedPermission).HandlerJSON(
		func(request *Request) any {
			uuid := request.Param("uuid")
			figure, err := app.getDashboardFigureData(uuid, request.user)
			must(err)
			return figure
			//request.app.templates.templates.ExecuteTemplate(request.Response(), "admin_form_table", table.TemplateData())
		},
	)
}

func (app *App) getDashboardTableData(uuid string, user *user) (*Table, error) {
	table := app.dashboardTableMap[uuid]
	if table == nil {
		return nil, errors.New("can't find table")
	}
	if !app.authorize(user, table.permission) {
		return nil, errors.New("can't authorize for access of table data")
	}
	return table.table(), nil
}

func (app *App) getDashboardFigureData(uuid string, user *user) (*DashboardFigureData, error) {
	figure := app.dashboardFigureMap[uuid]
	if figure == nil {
		return nil, errors.New("can't find figure")
	}
	if !app.authorize(user, figure.permission) {
		return nil, errors.New("can't authorize for access of figure data")
	}
	return figure.data(app), nil
}

func (board *Board) Dashboard(name func(string) string) *Dashboard {
	group := &Dashboard{
		board: board,
		name:  name,
	}
	//board.app.dashboardGroupMap[group.uuid] = group
	board.dashboardGroups = append(board.dashboardGroups, group)
	return group
}

func (group *Dashboard) Figure(name string, permission Permission) *DashboardFigure {
	must(group.board.app.validatePermission(permission))
	figure := &DashboardFigure{
		uuid:       randomString(30),
		name:       name,
		permission: permission,
	}
	group.board.app.dashboardFigureMap[figure.uuid] = figure
	group.figures = append(group.figures, figure)
	return figure
}

func (group *Dashboard) Table(tableFn func() *Table, permission Permission) *Dashboard {
	table := &dashboardTable{
		uuid:       randomString(30),
		table:      tableFn,
		permission: permission,
	}
	group.tables = append(group.tables, table)
	group.board.app.dashboardTableMap[table.uuid] = table
	return group
}

func (item *DashboardFigure) getValues(app *App) [2]int64 {
	var val, cmpVal int64 = -1, -1
	if item.value != nil {
		val = item.value()
	}
	if item.compareValue != nil {
		cmpVal = item.compareValue()
	}
	return [2]int64{
		val, cmpVal,
	}

}

func (item *DashboardFigure) Value(value func() int64) *DashboardFigure {
	item.value = value
	return item
}

func (item *DashboardFigure) Compare(value func() int64, description string) *DashboardFigure {
	item.compareValue = value
	item.compareDescription = description
	return item
}

func (item *DashboardFigure) getValueStr(values [2]int64, app *App) string {
	ret := "–"
	if item.value != nil {
		val := values[0]
		ret = humanizeNumber(val)
		if item.unit != "" {
			ret += " " + item.unit
		}
	}

	return ret
}

func (item *DashboardFigure) getDescriptionStr(values [2]int64, app *App) string {
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

	if item.unit != "" {
		ret += " " + item.unit
	}

	if item.compareDescription != "" {
		ret += " " + item.compareDescription
	}

	if compareValue > 0 {
		percent := fmt.Sprintf("%.2f%%", (100*float64(diff))/float64(compareValue))
		ret += fmt.Sprintf(" (%s)", percent)
	}
	return ret
}

func (item *DashboardFigure) view(app *App) *DashboardViewFigure {
	ret := &DashboardViewFigure{
		UUID: item.uuid,
		URL:  item.url,
		Name: item.name,
	}
	return ret
}

func (item *DashboardFigure) getColors(values [2]int64, app *App) (isGreen, isRed bool) {
	val, compareValue := values[0], values[1] //item.getValues(app)
	if val > compareValue {
		isGreen = true
	}
	if val < compareValue {
		isRed = true
	}
	return
}

func (item *DashboardFigure) URL(url string) *DashboardFigure {
	item.url = url
	return item
}

func (item *DashboardFigure) Unit(unit string) *DashboardFigure {
	item.unit = unit
	return item
}
