package prago

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

type FormTaskActivity struct {
	mutex            *sync.RWMutex
	uuid             string
	description      string
	progress         float64
	finished         bool
	stoppedByUser    bool
	tableRows        []*tableRow
	lastStateRequest time.Time
	ctx              context.Context
	cancel           context.CancelFunc
	isError          bool
}

func (fta *FormTaskActivity) Context() context.Context {
	if fta == nil {
		return context.Background()
	}
	return fta.ctx
}

func (fta *FormTaskActivity) checkIfStop() {

	if fta.lastStateRequest.Before(time.Now().Add(-2 * time.Minute)) {
		fta.isError = true
		fta.stoppedByUser = true
	}

	if fta.stoppedByUser {
		panic("stopped by user")
	}

}

func (fta *FormTaskActivity) Progress(finishedSoFar, total int64) {
	if fta == nil {
		return
	}

	fta.mutex.Lock()
	defer fta.mutex.Unlock()

	fta.checkIfStop()

	var progress float64 = 0
	if finishedSoFar > 0 && total > 0 {
		progress = float64(finishedSoFar) / float64(total)
	}
	fta.progress = progress
}

func (fta *FormTaskActivity) Description(description string) {
	if fta == nil {
		return
	}

	fta.mutex.Lock()
	defer fta.mutex.Unlock()

	fta.checkIfStop()

	fta.description = description
}

func (fta *FormTaskActivity) TableCells(cells ...*TableCell) {
	if fta == nil {
		return
	}

	fta.mutex.Lock()
	defer fta.mutex.Unlock()

	fta.checkIfStop()

	row := &tableRow{}
	row.Cells = cells

	fta.tableRows = append(fta.tableRows, row)
}

func newFormTaskActivity(request *Request, handler func(*FormTaskActivity) error) *FormTaskActivity {
	ctx, cancel := context.WithCancel(context.Background())
	ret := &FormTaskActivity{
		mutex:            &sync.RWMutex{},
		uuid:             randomString(30),
		lastStateRequest: time.Now(),
		ctx:              ctx,
		cancel:           cancel,
	}

	request.app.setFormTaskActivity(ret)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if ret.stoppedByUser {
					ret.description = "Ukončeno uživatelem"
				} else {
					recoveryStr := fmt.Sprintf("%v, stack: %s", r, string(debug.Stack()))
					ret.description = recoveryStr
					ret.isError = true
				}
			}

			request.app.taskActivityCleanup(ret.uuid)
			ret.finished = true
		}()
		err := handler(ret)
		if err != nil {
			ret.description = err.Error()
			ret.isError = true
		}
	}()

	return ret
}

func (app *App) getFormTaskActivity(uuid string) *FormTaskActivity {
	app.formTasksMutex.Lock()
	defer app.formTasksMutex.Unlock()
	return app.formTasksMap[uuid]
}

func (app *App) setFormTaskActivity(activity *FormTaskActivity) {
	app.formTasksMutex.Lock()
	defer app.formTasksMutex.Unlock()
	app.formTasksMap[activity.uuid] = activity
}

func (app *App) deleteFormTaskActivity(uuid string) {
	app.formTasksMutex.Lock()
	defer app.formTasksMutex.Unlock()
	delete(app.formTasksMap, uuid)
}

func (app *App) stopFormTask(uuid string) {
	task := app.getFormTaskActivity(uuid)
	task.stoppedByUser = true
	task.isError = true
	task.cancel()
	//TODO: wait till finished
	for {
		time.Sleep(200 * time.Millisecond)
		if task.finished {
			return
		}
	}
}

func (app *App) taskActivityCleanup(uuid string) {
	go func() {
		time.Sleep(1 * time.Minute)
		app.deleteFormTaskActivity(uuid)
	}()
}

func (app *App) initFormTask() {

	app.formTasksMap = map[string]*FormTaskActivity{}

	app.API("_taskview").HandlerJSON(func(request *Request) any {
		activity := app.getFormTaskActivity(request.Param("uuid"))
		if activity == nil {
			request.WriteJSON(404, "Not Found")
			return nil
		}
		ret := activity.toView()
		return ret
	}).Permission(everybodyPermission)

	app.API("_taskviewtable").Handler(func(request *Request) {
		activity := app.getFormTaskActivity(request.Param("uuid"))
		if activity == nil {
			request.WriteJSON(404, "Not Found")
			return
		}
		tableViewData := activity.getTableData()
		if len(tableViewData) == 0 {
			request.WriteJSON(204, "Empty")
			return
		}
		request.WriteHTML(200, app.adminTemplates, "table_rows", tableViewData)
	}).Permission(everybodyPermission)

	PopupForm(app, "_taskstop", func(form *Form, request *Request) {
		form.AddHidden("uuid").Value = request.Param("uuid")
		form.AutosubmitFirstTime = true
	}, func(fv FormValidation, request *Request) {
		app.stopFormTask(request.Param("uuid"))
		task := app.getFormTaskActivity(request.Param("uuid"))
		fv.Data(task.toView())
	}).Permission(everybodyPermission).Name(unlocalized("Ukončuji úlohu…"))

	ActionForm(app, "form-task-example", func(form *Form, request *Request) {
		form.AddSubmit("Spustit")
	}, func(fv FormValidation, request *Request) {

		fv.RunTask(request, func(fta *FormTaskActivity) error {
			for i := range 5 {
				fta.Description(fmt.Sprintf("Line %d dj iowqj dwioq jdiwoqj diwoqj diwoqj diwoqjdiwoqjdiwoq jdiwoq jdiwq", i))
				fta.Progress(int64(i)+1, 10)
				fta.TableCells(Cell("XXX"), Cell(i))
				time.Sleep(1000 * time.Millisecond)
			}
			fta.Description("Dokončeno")
			return nil
		})

	}).Permission("sysadmin").Name(unlocalized("Form úloha vzor")).Board(sysadminBoard)

}
