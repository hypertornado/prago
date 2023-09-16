package prago

import (
	"fmt"
	"io"
	"mime/multipart"
	"runtime/debug"
	"sort"
	"time"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type taskManager struct {
	app       *App
	tasksMap  map[string]*Task
	startedAt time.Time
}

type taskViewData struct {
	Title  string
	Locale string
	Tasks  []taskViewGroup
}

func GetTaskViewData(request *Request) taskViewData {
	var ret taskViewData
	userID := request.UserID()
	ret.Locale = request.Locale()
	csrfToken := request.app.generateCSRFToken(userID)
	ret.Tasks = request.app.taskManager.getTasks(userID, request, csrfToken)
	ret.Title = messages.Get(request.Locale(), "tasks")
	return ret
}

func (app *App) preInitTaskManager() {
	app.taskManager = &taskManager{
		app:       app,
		tasksMap:  make(map[string]*Task),
		startedAt: time.Now(),
	}

}

func (app *App) postInitTaskManager() {
	go app.taskManager.startCRON()

	//app.Action("tasks").Permission(loggedPermission).Name(messages.GetNameFunction("tasks")).Template("admin_tasks").DataSource(GetTaskViewData)

	app.API("tasks/runtask").Method("POST").Permission(loggedPermission).Handler(func(request *Request) {
		id := request.Request().FormValue("id")
		csrf := request.Request().FormValue("csrf")

		if request.csrfToken() != csrf {
			panic("wrong token")
		}

		task := app.taskManager.tasksMap[id]
		if !request.Authorize(task.permission) {
			panic("not authorize")
		}
		app.taskManager.run(task, request.UserID(), request.Locale(), request.Request().MultipartForm)
		request.Redirect("/admin")
	})

	app.sysadminTaskGroup = app.TaskGroup(unlocalized("Sysadmin"))

	app.sysadminTaskGroup.Task(unlocalized("Delete cache")).Handler(func(ta *TaskActivity) error {
		app.ClearCache()
		return nil
	})

	grp := app.TaskGroup(unlocalized("example"))

	grp.Task(unlocalized("example_simple ew oifeqio fjewqio fjeiwoq fjeioqwjf eiwoqf jeiowq")).Handler(func(t *TaskActivity) error {
		var progress float64
		for {
			time.Sleep(1000 * time.Millisecond)
			t.SetStatus(progress, "example status woiqfje iwoqfjeiwo qfjeiwoq jfeiowq fjeiw oqfjewioq")
			progress += 0.01
			if progress >= 1 {
				return nil
			}
		}
	})

	grp.Task(unlocalized("example_fail")).Handler(func(t *TaskActivity) error {
		return fmt.Errorf("example error")
	})

	grp.Task(unlocalized("example_panic")).Handler(func(t *TaskActivity) error {
		panic("panic value")
	})

	grp.Task(unlocalized("example")).FileInput("file_example").Handler(func(t *TaskActivity) error {
		file, err := t.GetFile("file_example")
		if err != nil {
			return err
		}

		data, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		fmt.Println(data)

		var progress float64
		for {
			time.Sleep(1000 * time.Millisecond)
			t.SetStatus(progress, "example status")
			progress += 0.01
			if progress >= 1 {
				return nil
			}
		}
	})
}

func (tm *taskManager) startCRON() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			for _, v := range tm.tasksMap {
				if v.cron > 0 {
					if tm.startedAt.Add(v.cron).Before(time.Now()) && v.lastStarted.Add(v.cron).Before(time.Now()) {
						tm.run(v, 0, "en", nil)
					}
				}
			}
		}
	}()
}

type taskViewGroup struct {
	Name  string
	Tasks []taskView
}

type taskView struct {
	ID        string
	Name      string
	CSRFToken string
	Files     []*taskFileInput
}

func (t *Task) taskView(locale, csrfToken string) taskView {
	return taskView{
		ID:        t.id,
		Name:      t.name(locale),
		CSRFToken: csrfToken,
		Files:     t.files,
	}
}

func (tm *taskManager) getTasks(userID int64, userData UserData, csrfToken string) (ret []taskViewGroup) {

	var tasks []*Task
	for _, v := range tm.tasksMap {
		if userData.Authorize(v.permission) {
			tasks = append(tasks, v)
		}
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		t1 := tasks[i]
		t2 := tasks[j]

		compareGroup := collate.New(language.Czech).CompareString(
			t1.group.name(userData.Locale()),
			t2.group.name(userData.Locale()),
		)
		if compareGroup < 0 {
			return true
		}
		if compareGroup > 0 {
			return false
		}

		compareID := collate.New(language.Czech).CompareString(t1.id, t2.id)
		return compareID < 0
	})

	var lastGroup *TaskGroup
	for _, v := range tasks {
		if v.group != lastGroup {
			ret = append(ret, taskViewGroup{Name: v.group.name(userData.Locale())})
		}

		ret[len(ret)-1].Tasks = append(ret[len(ret)-1].Tasks, v.taskView(userData.Locale(), csrfToken))
		lastGroup = v.group
	}

	for _, v := range ret {
		sortTaskViews(v.Tasks)
	}

	return ret
}

func sortTaskViews(items []taskView) {
	collator := collate.New(language.Czech)
	sort.SliceStable(items, func(i, j int) bool {
		a := items[i]
		b := items[j]
		if collator.CompareString(a.Name, b.Name) <= 0 {
			return true
		} else {
			return false
		}
	})
}

// Task represent some user task
type Task struct {
	id          string
	name        func(string) string
	group       *TaskGroup
	permission  Permission
	handler     func(*TaskActivity) error
	cron        time.Duration
	lastStarted time.Time
	files       []*taskFileInput
}

type taskFileInput struct {
	ID string
}

// Task creates task
func (tg *TaskGroup) Task(name func(string) string) *Task {
	id := randomString(20)
	_, ok := tg.manager.tasksMap[id]

	if ok {
		panic(fmt.Sprintf("Task '%s' already added.", id))
	}

	task := &Task{
		id:         id,
		name:       name,
		permission: sysadminPermission,
		group:      tg,
	}

	tg.tasks = append(tg.tasks, task)
	tg.manager.tasksMap[task.id] = task

	return task
}

// Handler sets handler to task
func (t *Task) Handler(fn func(*TaskActivity) error) *Task {
	t.handler = fn
	return t
}

// FileInput
func (t *Task) FileInput(id string) *Task {
	t.files = append(t.files, &taskFileInput{
		ID: id,
	})
	return t
}

// SetPermission set permission to task
func (t *Task) Permission(permission string) *Task {
	t.permission = Permission(permission)
	return t
}

// RepeatEvery sets cron to task
func (t *Task) RepeatEvery(duration time.Duration) *Task {
	t.cron = duration
	return t
}

// TaskGroup represent group of tasks
type TaskGroup struct {
	name    func(string) string
	manager *taskManager
	tasks   []*Task
}

// NewTaskGroup creates new task group
func (app *App) TaskGroup(name func(string) string) *TaskGroup {
	return &TaskGroup{
		name:    name,
		manager: app.taskManager,
	}
}

func (tm *taskManager) run(t *Task, userID int64, locale string, form *multipart.Form) {

	var name string = t.name(locale)

	var notification *Notification = tm.app.Notification(name)
	notification.preName = t.group.name(locale)

	activity := &TaskActivity{
		task:         t,
		notification: notification,
		files:        form,
	}
	t.lastStarted = time.Now()

	notification.SetPrimaryAction("Ukončit", func() {
		activity.stoppedByUser = true
	})

	notification.disableCancel = true

	if userID > 0 {
		notification.Push(userID)
	}

	progress := -1.0
	notification.SetProgress(&progress)

	go func() {
		defer func() {
			notification.primaryAction = nil
			notification.secondaryAction = nil
			activity.notification.disableCancel = false
			activity.notification.progress = nil
			if r := recover(); r != nil {
				activity.notification.SetStyleFail()
				if activity.stoppedByUser {
					notification.SetDescription("Ukončeno uživatelem")
				} else {
					recoveryStr := fmt.Sprintf("%v, stack: %s", r, string(debug.Stack()))
					notification.SetDescription(recoveryStr)
				}
			} else {
				activity.notification.description = "Úspěšně dokončeno"
				activity.notification.SetStyleSuccess()
			}
		}()
		must(t.handler(activity))
	}()
}
