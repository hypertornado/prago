package prago

import "errors"

type dashboardTable struct {
	uuid               string
	table              func(*Request) *Table
	permission         Permission
	refreshTimeSeconds int64
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

func (group *Dashboard) Table(tableFn func(*Request) *Table, permission Permission) *dashboardTable {
	table := &dashboardTable{
		uuid:       randomString(30),
		table:      tableFn,
		permission: permission,
	}
	group.tables = append(group.tables, table)
	group.board.app.dashboardTableMap[table.uuid] = table
	table.refreshTimeSeconds = 300
	return table
}

func (table *dashboardTable) RefreshTime(seconds int64) *dashboardTable {
	if seconds < 1 {
		seconds = 1
	}
	table.refreshTimeSeconds = seconds
	return table
}
