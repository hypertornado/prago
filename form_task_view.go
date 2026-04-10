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
	TableRows    template.HTML
}

func (fta *FormTaskActivity) toView(app *App) *FormTaskView {
	fta.mutex.Lock()
	defer fta.mutex.Unlock()

	tableRows := fta.tableRows

	fta.tableRows = nil

	var rowsHTML template.HTML = ""

	for _, row := range tableRows {
		rowsHTML += app.adminTemplates.ExecuteToHTML("table_row", tableCellsToTableRowView(row))
	}
	fmt.Println("ROOWS", rowsHTML)

	ret := &FormTaskView{
		Description:  fta.description,
		Progress:     fta.progress,
		ProgressText: formatProgressText(fta.progress),
		Finished:     fta.finished,
		TableRows:    rowsHTML,
	}

	fta.lastStateRequest = time.Now()

	return ret
}

func formatProgressText(progress float64) string {
	s := fmt.Sprintf("%.1f %%", progress*100)
	return strings.ReplaceAll(s, ".", ",")
}
