package prago

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
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
	user := request.user
	ret.Locale = user.Locale
	csrfToken := request.app.generateCSRFToken(user)
	ret.Tasks = request.app.taskManager.getTasks(user, csrfToken)
	ret.Title = messages.Get(user.Locale, "tasks")
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
		if !app.authorize(request.user, task.permission) {
			panic("not authorize")
		}
		app.taskManager.run(task, request.user, request.Request().MultipartForm)
		request.Redirect("/admin")
	})

	app.sysadminTaskGroup = app.TaskGroup(unlocalized("Sysadmin"))

	grp := app.TaskGroup(unlocalized("example"))

	grp.Task(unlocalized("example_simple")).Handler(func(t *TaskActivity) error {
		var progress float64
		for {
			time.Sleep(1 * time.Second)
			t.SetStatus(progress, "example status")
			progress += 0.2
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

		data, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		fmt.Println(data)

		var progress float64
		for {
			time.Sleep(1 * time.Second)
			t.SetStatus(progress, "example status")
			progress += 0.2
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
						tm.run(v, nil, nil)
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

func (tm *taskManager) getTasks(user *user, csrfToken string) (ret []taskViewGroup) {

	var tasks []*Task
	for _, v := range tm.tasksMap {
		if tm.app.authorize(user, v.permission) {
			tasks = append(tasks, v)
		}
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		t1 := tasks[i]
		t2 := tasks[j]

		compareGroup := collate.New(language.Czech).CompareString(
			t1.group.name(user.Locale),
			t2.group.name(user.Locale),
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
			ret = append(ret, taskViewGroup{Name: v.group.name(user.Locale)})
		}

		ret[len(ret)-1].Tasks = append(ret[len(ret)-1].Tasks, v.taskView(user.Locale, csrfToken))
		lastGroup = v.group
	}

	return ret
}

//Task represent some user task
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

//Task creates task
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

	/*tg.manager.app.addCommand("task", id).Callback(func() {
		tg.manager.run(task, nil, nil)
	})*/

	return task
}

//Handler sets handler to task
func (t *Task) Handler(fn func(*TaskActivity) error) *Task {
	t.handler = fn
	return t
}

//FileInput
func (t *Task) FileInput(id string) *Task {
	t.files = append(t.files, &taskFileInput{
		ID: id,
	})
	return t
}

//SetPermission set permission to task
func (t *Task) Permission(permission string) *Task {
	t.permission = Permission(permission)
	return t
}

//RepeatEvery sets cron to task
func (t *Task) RepeatEvery(duration time.Duration) *Task {
	t.cron = duration
	return t
}

//TaskGroup represent group of tasks
type TaskGroup struct {
	name    func(string) string
	manager *taskManager
	tasks   []*Task
}

//NewTaskGroup creates new task group
func (app *App) TaskGroup(name func(string) string) *TaskGroup {
	return &TaskGroup{
		name:    name,
		manager: app.taskManager,
	}
}

func (tm *taskManager) run(t *Task, user *user, form *multipart.Form) {
	var language = "en"
	if user != nil {
		language = user.Locale
	}

	var name string
	if user != nil {
		name = t.name(user.Locale)
	} else {
		name = t.name("en")
	}

	var notification *Notification = tm.app.Notification(name)
	notification.preName = t.group.name(language)

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

	if user != nil {
		notification.Push(user)
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
					recoveryStr := fmt.Sprintf("%v", r)
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
