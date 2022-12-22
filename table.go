package prago

import (
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/tealeg/xlsx"
)

type Table struct {
	app  *App
	data []*tableData
}

type tableData struct {
	Rows       []*tableRow
	FooterText []string
}

type tableRow struct {
	Cells []*TableCell
}

type TableCell struct {
	CSSClasses []string
	Href       string
	Text       string
	//Content    string
}

func (app *App) Table() *Table {
	ret := &Table{
		app: app,
	}
	ret.Table()
	return ret
}

func (table *Table) Table() *Table {
	table.data = append(table.data, &tableData{})
	table.newRow()
	return table
}

func (table *Table) newRow() {
	table.currentTable().Rows = append(table.currentTable().Rows, &tableRow{})

}

func (table *Table) Header(items ...interface{}) *Table {
	for _, item := range items {
		cell := table.Cell(item)
		cell.Header()
	}
	table.newRow()
	return table
}

func (table *Table) Row(items ...interface{}) *Table {
	for _, item := range items {
		table.Cell(item)
	}
	table.newRow()
	return table
}

func (t *Table) Cell(data interface{}) *TableCell {
	row := t.currentRow()
	cell := newCell(data)
	row.Cells = append(row.Cells, cell)
	return cell
}

func (cell *TableCell) Header() *TableCell {
	cell.CSSClass("form_table_cell-header")
	return cell
}

func (cell *TableCell) Pre() *TableCell {
	cell.CSSClass("form_table_cell-pre")
	return cell
}

func (cell *TableCell) URL(link string) *TableCell {
	cell.Href = link
	return cell
}

func (cell *TableCell) CSSClass(class string) *TableCell {
	cell.CSSClasses = append(cell.CSSClasses, class)
	return cell
}

func (cell *TableCell) GetClassesString() template.CSS {
	return template.CSS(strings.Join(cell.CSSClasses, " "))
}

func (table *Table) currentTable() *tableData {
	return table.data[len(table.data)-1]
}

func (table *Table) currentRow() *tableRow {
	currentTable := table.currentTable()
	return currentTable.Rows[len(currentTable.Rows)-1]
}

func newCell(item interface{}) *TableCell {
	ret := &TableCell{
		Text: fmt.Sprintf("%v", item),
	}

	ret.CSSClass("form_table_cell")

	linkData, ok := item.([2]string)
	if ok {
		ret = newCell(linkData[1])
		ret.URL(linkData[0])

	}

	return ret
}

func (t *Table) AddFooterText(text string) {
	t.currentTable().FooterText = append(t.currentTable().FooterText, text)
}

// TODO execute right into
func (t *Table) ExecuteHTML() template.HTML {
	return template.HTML(
		t.app.ExecuteTemplateToString("form_table", t.TemplateData()),
	)
}

func (t *Table) TemplateData() []*tableData {
	return t.data
}

func (t *Table) ExportXLSX(writer io.Writer) error {
	f := xlsx.NewFile()

	for k, table := range t.data {
		sheet, err := f.AddSheet(fmt.Sprintf("Sheet %d", k+1))
		must(err)

		for _, v1 := range table.Rows {
			row := sheet.AddRow()
			for _, v2 := range v1.Cells {
				cell := row.AddCell()
				cell.SetValue(v2.Text)
			}
		}
	}

	return f.Write(writer)
}
