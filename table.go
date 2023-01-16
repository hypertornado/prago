package prago

import (
	"fmt"
	"html/template"
	"io"
	"reflect"
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
	data *tableCellData
}

type tableCellData struct {
	CSSClasses []string
	Href       string
	Text       string
	Colspan    int64
	Rowspan    int64
}

type tableView struct {
	Rows       []*tableRowView
	FooterText []string
}

type tableRowView struct {
	Cells []*tableCellData
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

func (cell *TableCell) Center() *TableCell {
	cell.CSSClass("form_table_cell-center")
	return cell
}

func (cell *TableCell) Right() *TableCell {
	cell.CSSClass("form_table_cell-right")
	return cell
}

func (cell *TableCell) Green() *TableCell {
	cell.CSSClass("form_table_cell-green")
	return cell
}

func (cell *TableCell) Nowrap() *TableCell {
	cell.CSSClass("form_table_cell-nowrap")
	return cell
}

func (cell *TableCell) URL(link string) *TableCell {
	cell.data.Href = link
	return cell
}

func (cell *TableCell) Colspan(i int64) *TableCell {
	cell.data.Colspan = i
	return cell
}

func (cell *TableCell) Rowspan(i int64) *TableCell {
	cell.data.Rowspan = i
	return cell
}

func (cell *TableCell) CSSClass(class string) *TableCell {
	cell.data.CSSClasses = append(cell.data.CSSClasses, class)
	return cell
}

func (cell *tableCellData) GetClassesString() template.CSS {
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

	var number bool

	if reflect.TypeOf(item).Kind() == reflect.Int || reflect.TypeOf(item).Kind() == reflect.Int64 {
		item = humanizeNumber(item.(int64))
		number = true
	}

	ret := &TableCell{
		data: &tableCellData{
			Text: fmt.Sprintf("%v", item),
		},
	}

	ret.CSSClass("form_table_cell")
	linkData, ok := item.([2]string)
	if ok {
		ret = newCell(linkData[1])
		ret.URL(linkData[0])

	}

	if number {
		ret.Right()
		ret.Nowrap()
	}

	return ret
}

func (t *Table) AddFooterText(text string) {
	t.currentTable().FooterText = append(t.currentTable().FooterText, text)
}

// TODO execute right into response
func (t *Table) ExecuteHTML() template.HTML {
	return template.HTML(
		t.app.ExecuteTemplateToString("form_table", t.templateData()),
	)
}

func (t *Table) templateData() []*tableView {
	var ret []*tableView

	for _, v := range t.data {

		view := &tableView{
			FooterText: v.FooterText,
		}

		for _, v2 := range v.Rows {
			row := &tableRowView{}

			for _, v3 := range v2.Cells {
				//cell := &tableCellData{}
				row.Cells = append(row.Cells, v3.data)
			}

			view.Rows = append(view.Rows, row)
		}

		ret = append(ret, view)
	}

	return ret
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
				cell.SetValue(v2.data.Text)
			}
		}
	}

	return f.Write(writer)
}
