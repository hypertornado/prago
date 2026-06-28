package prago

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

type FormTaskView struct {
	Description  string
	Progress     float64
	ProgressText string
	Finished     bool
	IsError      bool
}

func (fta *FormTaskActivity) toView() *FormTaskView {
	fta.mutex.Lock()
	defer fta.mutex.Unlock()

	ret := &FormTaskView{
		Description:  fta.description,
		Progress:     fta.progress,
		ProgressText: formatProgressText(fta.progress),
		Finished:     fta.finished,
		IsError:      fta.isError,
	}

	fta.lastStateRequest = time.Now()

	return ret
}

func (fta *FormTaskActivity) getTableRows() [][]*TableCell {
	fta.mutex.Lock()
	defer fta.mutex.Unlock()

	tableRows := fta.tableRows
	fta.tableRows = nil

	return tableRows

}

func (fta *FormTaskActivity) getTableData(app *App) template.HTML {
	tableRows := fta.getTableRows()

	if len(tableRows) == 0 {
		return ""
	}

	var rowsHTML template.HTML = ""
	for _, row := range tableRows {
		rowsHTML += app.adminTemplates.ExecuteToHTML("table_row", tableCellsToTableRowView(row))
	}

	return rowsHTML
}

func formatProgressText(progress float64) string {
	s := fmt.Sprintf("%.1f %%", progress*100)
	return strings.ReplaceAll(s, ".", ",")
}
