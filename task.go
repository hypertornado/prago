package prago

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/hypertornado/prago/messages"
	"github.com/hypertornado/prago/utils"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type taskManager struct {
	app           *App
	tasks         []*Task
	tasksMap      map[string]*Task
	activities    map[string]*TaskActivity
	activityMutex *sync.RWMutex
	startedAt     time.Time
}

func (app *App) initTaskManager() {
	tm := &taskManager{
		app:           app,
		tasksMap:      make(map[string]*Task),
		activities:    make(map[string]*TaskActivity),
		activityMutex: &sync.RWMutex{},
		startedAt:     time.Now(),
	}

	app.taskManager = tm
	app.taskManager.init()
}

func (tm *taskManager) startCRON() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			for _, v := range tm.tasks {
				if v.cron > 0 {
					if tm.startedAt.Add(v.cron).Before(time.Now()) && v.lastStarted.Add(v.cron).Before(time.Now()) {
						tm.run(v, nil, "cron")
					}
				}
			}
		}
	}()
}

func (tm *taskManager) oldTasksRemover() {
	for {
		time.Sleep(1 * time.Second)
		for _, v := range tm.getOldActivities() {
			tm.deleteActivity(v)
		}
	}

}

func (tm *taskManager) getOldActivities() (ret []string) {
	tm.activityMutex.RLock()
	defer tm.activityMutex.RUnlock()
	for k, v := range tm.activities {
		if v.ended && v.endedAt.Add(24*time.Hour).Before(time.Now()) {
			ret = append(ret, k)
		}
	}
	return ret
}

func (tm *taskManager) deleteActivity(id string) {
	tm.activityMutex.Lock()
	defer tm.activityMutex.Unlock()
	delete(tm.activities, id)
}

func (tm *taskManager) init() {
	go tm.oldTasksRemover()
	go tm.startCRON()

	tm.app.AdminController.Get(tm.app.GetAdminURL("_tasks"), func(request Request) {
		user := GetUser(request)
		request.SetData("tasks", tm.getTasks(user))
		request.SetData("taskmonitor", tm.getTaskMonitor(user))
		request.SetData("admin_yield", "admin_tasks")
		request.SetData("admin_title", messages.Messages.Get(user.Locale, "tasks"))
		request.RenderView("admin_layout")
	})

	tm.app.AdminController.Get(tm.app.GetAdminURL("_tasks/running"), func(request Request) {
		request.SetData("taskmonitor", tm.getTaskMonitor(GetUser(request)))
		request.RenderView("taskmonitor")
	})

	tm.app.AdminController.Post(tm.app.GetAdminURL("_tasks/runtask"), func(request Request) {
		id := request.Request().FormValue("id")
		csrf := request.Request().FormValue("csrf")
		user := GetUser(request)

		expectedToken := request.GetData("_csrfToken").(string)
		if expectedToken != csrf {
			panic("wrong token")
		}

		must(tm.startTask(id, user))
		request.Redirect(tm.app.GetAdminURL("_tasks"))
	})

	tm.app.AdminController.Get(tm.app.GetAdminURL("_tasks/stoptask"), func(request Request) {
		uuid := request.Request().FormValue("uuid")
		csrf := request.Request().FormValue("csrf")
		user := GetUser(request)

		expectedToken := request.GetData("_csrfToken").(string)
		if expectedToken != csrf {
			panic("wrong token")
		}

		must(tm.stopTask(uuid, user))
		request.Redirect(tm.app.GetAdminURL("_tasks"))
	})

	tm.app.AdminController.Get(tm.app.GetAdminURL("_tasks/deletetask"), func(request Request) {
		uuid := request.Request().FormValue("uuid")
		csrf := request.Request().FormValue("csrf")
		user := GetUser(request)

		expectedToken := request.GetData("_csrfToken").(string)
		if expectedToken != csrf {
			panic("wrong token")
		}

		must(tm.deleteTask(uuid, user))
		request.Redirect(tm.app.GetAdminURL("_tasks"))
	})

	grp := tm.app.NewTaskGroup(Unlocalized("example"))

	tm.app.NewTask("example_fail").SetGroup(grp).SetHandler(func(t *TaskActivity) error {
		return fmt.Errorf("example error")
	})

	tm.app.NewTask("example").SetGroup(grp).SetHandler(func(t *TaskActivity) error {
		t.IsStopped()
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

func (tm *taskManager) startTask(id string, user User) error {
	task, ok := tm.tasksMap[id]
	if !ok {
		return fmt.Errorf("Can't find task %s", id)
	}

	if tm.app.Authorize(user, task.permission) {
		tm.run(task, &user, "button")
		return nil
	} else {
		return fmt.Errorf("User is not authorized to run this task")
	}
}

func (tm *taskManager) stopTask(uuid string, user User) error {
	tm.activityMutex.Lock()
	defer tm.activityMutex.Unlock()

	activity := tm.activities[uuid]
	if activity.user.ID != user.ID {
		return fmt.Errorf("Wrong user")
	}
	activity.stopped = true
	return nil
}

func (tm *taskManager) deleteTask(uuid string, user User) error {
	tm.activityMutex.Lock()
	defer tm.activityMutex.Unlock()

	activity := tm.activities[uuid]
	if activity.user.ID != user.ID {
		return fmt.Errorf("Wrong user")
	}

	delete(tm.activities, uuid)
	return nil
}

type TaskViewGroup struct {
	Name  string
	Tasks []TaskView
}

type TaskView struct {
	ID   string
	Name string
}

func (t *Task) taskView() TaskView {
	return TaskView{
		ID:   t.id,
		Name: t.id,
	}
}

func (tm *taskManager) getTasks(user User) (ret []TaskViewGroup) {

	var tasks []*Task
	for _, v := range tm.tasks {
		if tm.app.Authorize(user, v.permission) {
			tasks = append(tasks, v)
		}
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		t1 := tasks[i]
		t2 := tasks[j]

		compareGroup := collate.New(language.Czech).CompareString(
			t1.group.Name(user.Locale),
			t2.group.Name(user.Locale),
		)
		if compareGroup < 0 {
			return true
		}
		if compareGroup > 0 {
			return false
		}

		compareID := collate.New(language.Czech).CompareString(t1.id, t2.id)
		if compareID < 0 {
			return true
		}
		return false
	})

	var lastGroup *taskGroup
	for _, v := range tasks {
		if v.group != lastGroup {
			ret = append(ret, TaskViewGroup{Name: v.group.Name(user.Locale)})
		}

		ret[len(ret)-1].Tasks = append(ret[len(ret)-1].Tasks, v.taskView())
		lastGroup = v.group
	}

	return ret
}

type Task struct {
	id          string
	group       *taskGroup
	permission  Permission
	handler     func(*TaskActivity) error
	cron        time.Duration
	lastStarted time.Time
	manager     *taskManager
}

var defaultGroup *taskGroup

func (app *App) NewTask(id string) *Task {
	if defaultGroup == nil {
		defaultGroup = app.NewTaskGroup(Unlocalized("Other"))
	}

	_, ok := app.taskManager.tasksMap[id]
	if ok {
		panic(fmt.Sprintf("Task '%s' already added.", id))
	}

	task := &Task{
		id:      id,
		manager: app.taskManager,
		group:   defaultGroup,
	}

	task.manager.tasks = append(app.taskManager.tasks, task)
	task.manager.tasksMap[task.id] = task

	app.AddCommand("task", id).Callback(func() {
		app.taskManager.run(task, nil, "command")
	})

	return task
}

func (t *Task) SetHandler(fn func(*TaskActivity) error) *Task {
	t.handler = fn
	return t
}

func (t *Task) SetPermission(permission string) *Task {
	t.permission = Permission(permission)
	return t
}

func (t *Task) RepeatEvery(duration time.Duration) *Task {
	t.cron = duration
	return t
}

type taskGroup struct {
	Name func(string) string
}

func (app *App) NewTaskGroup(name func(string) string) *taskGroup {
	return &taskGroup{name}
}

func (t *Task) SetGroup(group *taskGroup) *Task {
	if group != nil {
		t.group = group
	}
	return t
}

func (tm *taskManager) run(t *Task, user *User, starterTyp string) *TaskActivity {
	activity := &TaskActivity{
		uuid:      utils.RandomString(10),
		task:      t,
		user:      user,
		typ:       starterTyp,
		startedAt: time.Now(),
	}
	t.lastStarted = time.Now()
	tm.addActivity(activity)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				recoveryStr := fmt.Sprintf("Recovered in run task: %v", r)
				activity.SetStatus(1, recoveryStr)
			}
			activity.ended = true
			activity.endedAt = time.Now()
			if user != nil {
				err := tm.app.Notification(*user, "Task finished").Create()
				if err != nil {
					fmt.Println(err)
				}
			}
		}()
		if t.handler != nil {
			activity.error = t.handler(activity)
		}
	}()

	return activity
}
