package prago

import (
	"fmt"
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

func (fta *FormTaskActivity) getTableRows() []*tableRow {
	fta.mutex.Lock()
	defer fta.mutex.Unlock()

	tableRows := fta.tableRows
	fta.tableRows = nil

	return tableRows

}

func (fta *FormTaskActivity) getTableData() []*tableRowView {
	tableRows := fta.getTableRows()

	if len(tableRows) == 0 {
		return nil
	}

	var rowViews []*tableRowView
	for _, row := range tableRows {
		rowView := tableCellsToTableRowView(row.Cells)
		rowView.Reveal = true
		rowViews = append(rowViews, rowView)
	}

	return rowViews

}

func formatProgressText(progress float64) string {
	s := fmt.Sprintf("%.1f %%", progress*100)
	return strings.ReplaceAll(s, ".", ",")
}
