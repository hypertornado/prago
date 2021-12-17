package prago

import (
	"fmt"
)

type Table struct {
	data *tableData
}

type tableData struct {
	Rows  []tableRow
	Width int64
}

type tableRow struct {
	IsHeader bool
	Cells    []tableCell
}

type tableCell struct {
	Content string
}

func NewTable() *Table {
	return &Table{
		data: &tableData{},
	}
}

func newCell(item interface{}) tableCell {
	return tableCell{
		Content: fmt.Sprintf("%v", item),
	}
}

func (t *Table) tableRow(isHeader bool, items ...interface{}) {
	if t.data.Width < int64(len(items)) {
		t.data.Width = int64(len(items))
	}
	var row = tableRow{
		IsHeader: isHeader,
	}
	for _, v := range items {
		row.Cells = append(row.Cells, newCell(v))
	}
	t.data.Rows = append(t.data.Rows, row)
}

func (t *Table) Header(items ...interface{}) {
	t.tableRow(true, items...)
}

func (t *Table) Row(items ...interface{}) {
	t.tableRow(false, items...)
}

func (t *Table) TemplateData() *tableData {
	return t.data
}
