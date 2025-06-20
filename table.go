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
	Javascript template.JS
}

type tableRow struct {
	Cells []*TableCell
}

type TableCell struct {
	data *tableCellData
}

type tableCellData struct {
	CSSClasses        []string
	Href              string
	DescriptionBefore string
	Text              string
	DescriptionAfter  string
	TextAfter         string
	Colspan           int64
	Rowspan           int64
	Checkboxes        []*tableCellCheckbox
	Buttons           []*TableCellButton
}

type tableCellCheckbox struct {
	Name    string
	Checked bool
}

type TableCellButton struct {
	Name    string
	URL     string
	OnClick template.JS
}

type tableView struct {
	Rows       []*tableRowView
	FooterText []string
	Javascript template.JS
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

func (table *Table) Header(items ...string) *Table {
	var headerCells = []*TableCell{}
	for _, item := range items {
		cell := Cell(item)
		cell.Header()
		headerCells = append(headerCells, cell)
	}
	table.Row(headerCells...)
	//table.newRow()
	return table
}

func (table *Table) Javascript(javascript template.JS) *Table {
	table.data[0].Javascript = javascript
	return table
}

func (table *Table) Row(items ...*TableCell) *Table {
	row := table.currentRow()
	row.Cells = append(row.Cells, items...)
	table.newRow()
	return table
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

func (cell *TableCell) Orange() *TableCell {
	cell.CSSClass("form_table_cell-orange")
	return cell
}

func (cell *TableCell) Red() *TableCell {
	cell.CSSClass("form_table_cell-red")
	return cell
}

func (cell *TableCell) Nowrap() *TableCell {
	cell.CSSClass("form_table_cell-nowrap")
	return cell
}

func (cell *TableCell) DescriptionBefore(description string) *TableCell {
	cell.data.DescriptionBefore = description
	return cell
}

func (cell *TableCell) DescriptionAfter(description string) *TableCell {
	cell.data.DescriptionAfter = description
	return cell
}

func (cell *TableCell) TextAfter(text string) *TableCell {
	cell.data.TextAfter = text
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

func (cell *TableCell) Checkbox(name string, checked bool) *TableCell {
	cell.data.Checkboxes = append(cell.data.Checkboxes, &tableCellCheckbox{
		Name:    name,
		Checked: checked,
	})
	return cell
}

func (cell *TableCell) Button(btn *TableCellButton) *TableCell {
	cell.data.Buttons = append(cell.data.Buttons, btn)
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

func Cell(item interface{}) *TableCell {

	var number bool

	if reflect.TypeOf(item).Kind() == reflect.Int {
		item = humanizeNumber(int64(item.(int)))
		number = true
	}

	if reflect.TypeOf(item).Kind() == reflect.Int64 {
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
		ret = Cell(linkData[1])
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
	return t.app.adminTemplates.ExecuteToHTML("form_table", t.templateData())
}

func (t *Table) templateData() []*tableView {
	var ret []*tableView

	for _, v := range t.data {

		view := &tableView{
			FooterText: v.FooterText,
			Javascript: v.Javascript,
		}

		for _, v2 := range v.Rows {
			row := &tableRowView{}

			for _, v3 := range v2.Cells {
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
