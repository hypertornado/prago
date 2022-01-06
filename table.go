package prago

import (
	"fmt"
	"html/template"
)

type Table struct {
	app  *App
	data *tableData
}

type tableData struct {
	Rows       []tableRow
	FooterText []string
	Width      int64
}

type tableRow struct {
	IsHeader bool
	Cells    []tableCell
}

type tableCell struct {
	Content string
}

func (app *App) Table() *Table {
	return &Table{
		app:  app,
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

func (t *Table) AddFooterText(text string) {
	t.data.FooterText = append(t.data.FooterText, text)
}

func (t *Table) ExecuteHTML() template.HTML {
	return template.HTML(
		t.app.ExecuteTemplateToString("admin_form_table", t.TemplateData()),
	)
}

func (t *Table) TemplateData() *tableData {
	return t.data
}
