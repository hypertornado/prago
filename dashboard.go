package prago

import (
	"errors"
	"fmt"
)

type DashboardGroup struct {
	app   *App
	uuid  string
	name  func(string) string
	items []*DashboardItem
	table *dashboardGroupTable
}

type DashboardItem struct {
	uuid               string
	permission         Permission
	url                string
	value              func() int64
	compareValue       func() int64
	compareDescription string
	unit               string
	name               string
}

type dashboardGroupTable struct {
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

	for _, group := range app.dashboardGroups {
		if group.uuid == uuid {
			if !app.authorize(user, group.table.permission) {
				return nil, errors.New("can't authorize for access of table data")
			}
			return group.table.table(), nil
		}
	}
	return nil, errors.New("can't find table")

}

func (app *App) DashboardGroup(name func(string) string) *DashboardGroup {
	group := &DashboardGroup{
		app:  app,
		uuid: randomString(30),
		name: name,
	}
	app.dashboardGroups = append(app.dashboardGroups, group)
	return group
}

func (group *DashboardGroup) Item(name string, permission Permission) *DashboardItem {
	must(group.app.validatePermission(permission))
	item := &DashboardItem{
		uuid:       randomString(30),
		name:       name,
		permission: permission,
	}
	group.items = append(group.items, item)
	return item
}

func (group *DashboardGroup) Table(table func() *Table, permission Permission) *DashboardGroup {
	group.table = &dashboardGroupTable{
		table:      table,
		permission: permission,
	}
	return group
}

func (group *DashboardGroup) getTable() *Table {
	if group.table == nil {
		return nil
	}
	return <-Cached(group.app, fmt.Sprintf("dashboard-table-%s", group.uuid), func() *Table {
		return group.table.table()
	})
}

func (item *DashboardItem) getValues(app *App) (int64, int64) {
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

func (item *DashboardItem) Value(value func() int64) *DashboardItem {
	item.value = value
	return item
}

func (item *DashboardItem) Compare(value func() int64, description string) *DashboardItem {
	item.compareValue = value
	item.compareDescription = description
	return item
}

func (item *DashboardItem) getValueStr(app *App) string {
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

func (item *DashboardItem) getDescriptionStr(app *App) string {
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

func (item *DashboardItem) homeItem(app *App) *HomeItem {
	ret := &HomeItem{
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

func (item *DashboardItem) URL(url string) *DashboardItem {
	item.url = url
	return item
}

func (item *DashboardItem) Unit(unit string) *DashboardItem {
	item.unit = unit
	return item
}
