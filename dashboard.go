package prago

import (
	"errors"
	"fmt"
)

type Dashboard struct {
	board   *Board
	name    func(string) string
	figures []*DashboardFigure
	tables  []*DashboardTable
}

type DashboardFigure struct {
	uuid               string
	permission         Permission
	url                string
	value              func(*Request) int64
	compareValue       func(*Request) int64
	compareDescription func(string) string
	unit               func(string) string
	name               func(string) string
	refreshTimeSeconds int64
}

type DashboardTable struct {
	uuid               string
	table              func(*Request) *Table
	permission         Permission
	refreshTimeSeconds int64
}

type DashboardFigureData struct {
	Value       string
	Description string
	IsGreen     bool
	IsRed       bool
}

func (figure DashboardFigure) data(request *Request, app *App) *DashboardFigureData {
	values := figure.getValues(request, app)
	ret := &DashboardFigureData{
		Value:       figure.getValueStr(request, values, app),
		Description: figure.getDescriptionStr(request, values, app),
	}
	ret.IsGreen, ret.IsRed = figure.getColors(values, app)
	return ret
}

func (app *App) initDashboard() {

	app.API("dashboard-table").Method("GET").Permission(loggedPermission).Handler(
		func(request *Request) {
			uuid := request.Param("uuid")
			table, err := app.getDashboardTableData(request, uuid)
			must(err)
			request.app.templates.templates.ExecuteTemplate(request.Response(), "form_table", table.templateData())
		},
	)

	app.API("dashboard-figure").Method("GET").Permission(loggedPermission).HandlerJSON(
		func(request *Request) any {
			uuid := request.Param("uuid")
			figure, err := app.getDashboardFigureData(request, uuid)
			must(err)
			return figure
		},
	)
}

func (app *App) getDashboardTableData(request *Request, uuid string) (*Table, error) {
	table := app.dashboardTableMap[uuid]
	if table == nil {
		return nil, errors.New("can't find table")
	}
	if !request.Authorize(table.permission) {
		return nil, errors.New("can't authorize for access of table data")
	}
	return table.table(request), nil
}

func (app *App) getDashboardFigureData(request *Request, uuid string) (*DashboardFigureData, error) {
	figure := app.dashboardFigureMap[uuid]
	if figure == nil {
		return nil, errors.New("can't find figure")
	}
	if !request.Authorize(figure.permission) {
		return nil, errors.New("can't authorize for access of figure data")
	}
	return figure.data(request, app), nil
}

func (board *Board) Dashboard(name func(string) string) *Dashboard {
	group := &Dashboard{
		board: board,
		name:  name,
	}
	board.dashboardGroups = append(board.dashboardGroups, group)
	return group
}

func (dashboard *Dashboard) Figure(name func(string) string, permission Permission) *DashboardFigure {
	must(dashboard.board.app.validatePermission(permission))
	figure := &DashboardFigure{
		uuid:               randomString(30),
		name:               name,
		permission:         permission,
		refreshTimeSeconds: 60,
	}
	dashboard.board.app.dashboardFigureMap[figure.uuid] = figure
	dashboard.figures = append(dashboard.figures, figure)
	return figure
}

func (group *Dashboard) Table(tableFn func(*Request) *Table, permission Permission) *DashboardTable {
	table := &DashboardTable{
		uuid:       randomString(30),
		table:      tableFn,
		permission: permission,
	}
	group.tables = append(group.tables, table)
	group.board.app.dashboardTableMap[table.uuid] = table
	table.refreshTimeSeconds = 300
	return table
}

func (figure *DashboardFigure) RefreshTime(seconds int64) *DashboardFigure {
	if seconds < 1 {
		seconds = 1
	}
	figure.refreshTimeSeconds = seconds
	return figure
}

func (table *DashboardTable) RefreshTime(seconds int64) *DashboardTable {
	if seconds < 1 {
		seconds = 1
	}
	table.refreshTimeSeconds = seconds
	return table
}

func (group *Dashboard) isVisible(userData UserData) bool {
	for _, v := range group.figures {
		if userData.Authorize(v.permission) {
			return true
		}
	}

	for _, v := range group.tables {
		if userData.Authorize(v.permission) {
			return true
		}
	}
	return false
}

func (item *DashboardFigure) getValues(request *Request, app *App) [2]int64 {
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

func (item *DashboardFigure) Value(value func(*Request) int64) *DashboardFigure {
	item.value = value
	return item
}

func (item *DashboardFigure) Compare(value func(*Request) int64, description func(string) string) *DashboardFigure {
	item.compareValue = value
	item.compareDescription = description
	return item
}

func (item *DashboardFigure) getValueStr(request *Request, values [2]int64, app *App) string {
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

func (item *DashboardFigure) getDescriptionStr(request *Request, values [2]int64, app *App) string {
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

func (item *DashboardFigure) view(request *Request) *DashboardViewFigure {
	ret := &DashboardViewFigure{
		UUID:               item.uuid,
		URL:                item.url,
		Name:               item.name(request.Locale()),
		RefreshTimeSeconds: item.refreshTimeSeconds,
	}
	return ret
}

func (item *DashboardFigure) getColors(values [2]int64, app *App) (isGreen, isRed bool) {
	val, compareValue := values[0], values[1]
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

func (item *DashboardFigure) Unit(unit func(string) string) *DashboardFigure {
	item.unit = unit
	return item
}
