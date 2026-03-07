package prago

import (
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
	tableRows        [][]*TableCell
	lastStateRequest time.Time
}

func (fta *FormTaskActivity) checkIfStop() {

	if fta.lastStateRequest.Before(time.Now().Add(-20 * time.Second)) {
		fta.stoppedByUser = true
	}

	if fta.stoppedByUser {
		panic("stopped by user")
	}

}

func (fta *FormTaskActivity) Progress(finishedSoFar, total int64) {
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
	fta.mutex.Lock()
	defer fta.mutex.Unlock()

	fta.checkIfStop()

	fta.description = description
}

func (fta *FormTaskActivity) TableCells(cells ...*TableCell) {
	fta.mutex.Lock()
	defer fta.mutex.Unlock()

	fta.checkIfStop()

	fta.tableRows = append(fta.tableRows, cells)
}

func newFormTaskActivity(request *Request, handler func(*FormTaskActivity) error) *FormTaskActivity {
	ret := &FormTaskActivity{
		mutex:            &sync.RWMutex{},
		uuid:             randomString(30),
		lastStateRequest: time.Now(),
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
				}
			}

			request.app.taskActivityCleanup(ret.uuid)
			ret.finished = true
		}()
		must(handler(ret))
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
	for {
		time.Sleep(200 * time.Millisecond)
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

	app.API("_taskview").HandlerJSON(func(request *Request) interface{} {
		activity := app.getFormTaskActivity(request.Param("uuid"))
		if activity == nil {
			request.WriteJSON(404, "Not Found")
			return nil
		}
		return activity.toView(app)
	}).Permission(everybodyPermission)

	PopupForm(app, "_taskstop", func(form *Form, request *Request) {
		form.AddHidden("uuid").Value = request.Param("uuid")
		form.AutosubmitFirstTime = true
	}, func(fv FormValidation, request *Request) {
		app.stopFormTask(request.Param("uuid"))
		fv.Data("ok")
	}).Permission(everybodyPermission).Name(unlocalized("Ukončit úlohu"))

	ActionForm(app, "form-task-example", func(form *Form, request *Request) {
		form.AddSubmit("Spustit")
	}, func(fv FormValidation, request *Request) {

		fv.RunTask(request, func(fta *FormTaskActivity) error {
			for i := range 10 {
				fta.Description(fmt.Sprintf("Line %d", i))
				fta.Progress(int64(i)+1, 10)
				fta.TableCells(Cell("XXX"), Cell(i))
				time.Sleep(1000 * time.Millisecond)
			}
			fta.Description("Dokončeno")
			return nil
		})

	}).Permission("sysadmin").Name(unlocalized("Form úloha vzor")).Board(sysadminBoard)

}
