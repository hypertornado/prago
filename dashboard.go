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

func (app *App) initDashboard() {

	app.API("dashboard-table").Method("GET").Permission(loggedPermission).Handler(
		func(request *Request) {
			uuid := request.Param("uuid")
			table, err := app.getDashboardTableData(uuid, request.user)
			must(err)
			request.app.templates.templates.ExecuteTemplate(request.Response(), "admin_form_table", table.TemplateData())
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

func (board *Board) Dashboard(name func(string) string) *Dashboard {
	group := &Dashboard{
		board: board,
		name:  name,
	}
	//board.app.dashboardGroupMap[group.uuid] = group
	board.dashboardGroups = append(board.dashboardGroups, group)
	return group
}

func (group *Dashboard) Item(name string, permission Permission) *DashboardFigure {
	must(group.board.app.validatePermission(permission))
	item := &DashboardFigure{
		uuid:       randomString(30),
		name:       name,
		permission: permission,
	}
	group.figures = append(group.figures, item)
	return item
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

/*func (group *Dashboard) getTable() *Table {
	if group.table == nil {
		return nil
	}
	return <-Cached(group.board.app, fmt.Sprintf("dashboard-table-%s", group.uuid), func() *Table {
		return group.table.table()
	})
}*/

func (item *DashboardFigure) getValues(app *App) (int64, int64) {
	val := Cached(app, fmt.Sprintf("dashboard-value-%s", item.uuid), func() int64 {
		if item.value != nil {
			return item.value()
		}
		return -1
	})
	cmpVal := Cached(app, fmt.Sprintf("dashboard-comparevalue-%s", item.uuid), func() int64 {
		if item.compareValue != nil {
			return item.compareValue()
		}
		return -1
	})
	return <-val, <-cmpVal
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

func (item *DashboardFigure) getValueStr(app *App) string {
	ret := "–"
	if item.value != nil {
		val, _ := item.getValues(app)
		ret = humanizeNumber(val)
		if item.unit != "" {
			ret += " " + item.unit
		}
	}

	return ret
}

func (item *DashboardFigure) getDescriptionStr(app *App) string {
	if item.value == nil {
		return ""
	}
	if item.compareValue == nil {
		return ""
	}

	val, compareValue := item.getValues(app)

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

func (item *DashboardFigure) homeItem(app *App) *BoardViewItem {
	ret := &BoardViewItem{
		URL:         item.url,
		Name:        item.name,
		Value:       item.getValueStr(app),
		Description: item.getDescriptionStr(app),
	}

	val, compareValue := item.getValues(app)
	if val > compareValue {
		ret.IsGreen = true
	}
	if val < compareValue {
		ret.IsRed = true
	}
	return ret
}

func (item *DashboardFigure) URL(url string) *DashboardFigure {
	item.url = url
	return item
}

func (item *DashboardFigure) Unit(unit string) *DashboardFigure {
	item.unit = unit
	return item
}
