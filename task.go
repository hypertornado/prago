package prago

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/hypertornado/prago/utils"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type taskManager struct {
	app           *App
	taskGroups    []*TaskGroup
	tasksMap      map[string]*Task
	activities    map[string]*TaskActivity
	activityMutex *sync.RWMutex
	startedAt     time.Time
	defaultGroup  *TaskGroup
}

func (app *App) initTaskManager() {

	app.taskManager = &taskManager{
		app:           app,
		tasksMap:      make(map[string]*Task),
		activities:    make(map[string]*TaskActivity),
		activityMutex: &sync.RWMutex{},
		startedAt:     time.Now(),
	}
	app.taskManager.defaultGroup = app.TaskGroup(Unlocalized("Other"))

	go app.taskManager.oldTasksRemover()
	go app.taskManager.startCRON()

	app.Action("_tasks").Name(messages.GetNameFunction("tasks")).IsWide().Template("admin_tasks").DataSource(
		func(request *Request) interface{} {
			var ret = map[string]interface{}{}
			user := request.user
			ret["locale"] = user.Locale
			ret["csrf_token"] = app.generateCSRFToken(user)
			ret["tasks"] = app.taskManager.getTasks(user)
			ret["taskmonitor"] = app.taskManager.getTaskMonitor(user)
			ret["admin_title"] = messages.Get(user.Locale, "tasks")
			return ret
		},
	)

	app.adminController.get(app.getAdminURL("_tasks/running"), func(request *Request) {
		request.SetData("taskmonitor", app.taskManager.getTaskMonitor(request.user))
		request.RenderView("taskmonitor")
	})

	app.adminController.post(app.getAdminURL("_tasks/runtask"), func(request *Request) {
		id := request.Request().FormValue("id")
		csrf := request.Request().FormValue("csrf")

		expectedToken := request.GetData("_csrfToken").(string)
		if expectedToken != csrf {
			panic("wrong token")
		}

		must(app.taskManager.startTask(id, request.user))
		request.Redirect(app.getAdminURL("_tasks"))
	})

	app.adminController.get(app.getAdminURL("_tasks/stoptask"), func(request *Request) {
		uuid := request.Request().FormValue("uuid")
		csrf := request.Request().FormValue("csrf")

		expectedToken := request.GetData("_csrfToken").(string)
		if expectedToken != csrf {
			panic("wrong token")
		}

		must(app.taskManager.stopTask(uuid, request.user))
		request.Redirect(app.getAdminURL("_tasks"))
	})

	app.adminController.get(app.getAdminURL("_tasks/deletetask"), func(request *Request) {
		uuid := request.Request().FormValue("uuid")
		csrf := request.Request().FormValue("csrf")

		expectedToken := request.GetData("_csrfToken").(string)
		if expectedToken != csrf {
			panic("wrong token")
		}

		must(app.taskManager.deleteTask(uuid, request.user))
		request.Redirect(app.getAdminURL("_tasks"))
	})

	grp := app.TaskGroup(Unlocalized("example"))

	grp.Task("example_fail").Handler(func(t *TaskActivity) error {
		return fmt.Errorf("example error")
	})

	grp.Task("example").Handler(func(t *TaskActivity) error {
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

func (tm *taskManager) startCRON() {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			for _, v := range tm.tasksMap {
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

func (tm *taskManager) startTask(id string, user *User) error {
	task, ok := tm.tasksMap[id]
	if !ok {
		return fmt.Errorf("Can't find task %s", id)
	}

	if tm.app.authorize(user, task.permission) {
		tm.run(task, user, "button")
		return nil
	}
	return fmt.Errorf("User is not authorized to run this task")
}

func (tm *taskManager) stopTask(uuid string, user *User) error {
	tm.activityMutex.Lock()
	defer tm.activityMutex.Unlock()

	activity := tm.activities[uuid]
	if activity.user.ID != user.ID {
		return fmt.Errorf("Wrong user")
	}
	activity.stopped = true
	return nil
}

func (tm *taskManager) deleteTask(uuid string, user *User) error {
	tm.activityMutex.Lock()
	defer tm.activityMutex.Unlock()

	activity := tm.activities[uuid]
	if activity.user.ID != user.ID {
		return fmt.Errorf("Wrong user")
	}

	delete(tm.activities, uuid)
	return nil
}

type taskViewGroup struct {
	Name  string
	Tasks []taskView
}

type taskView struct {
	ID   string
	Name string
}

func (t *Task) taskView() taskView {
	return taskView{
		ID:   t.id,
		Name: t.id,
	}
}

func (tm *taskManager) getTasks(user *User) (ret []taskViewGroup) {

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
		if compareID < 0 {
			return true
		}
		return false
	})

	var lastGroup *TaskGroup
	for _, v := range tasks {
		if v.group != lastGroup {
			ret = append(ret, taskViewGroup{Name: v.group.name(user.Locale)})
		}

		ret[len(ret)-1].Tasks = append(ret[len(ret)-1].Tasks, v.taskView())
		lastGroup = v.group
	}

	return ret
}

//Task represent some user task
type Task struct {
	id          string
	group       *TaskGroup
	permission  Permission
	handler     func(*TaskActivity) error
	cron        time.Duration
	lastStarted time.Time
	//manager     *taskManager
}

var defaultGroup *TaskGroup

//Task creates task
func (tg *TaskGroup) Task(id string) *Task {
	_, ok := tg.manager.tasksMap[id]

	if ok {
		panic(fmt.Sprintf("Task '%s' already added.", id))
	}

	task := &Task{
		id:    id,
		group: tg,
	}

	tg.tasks = append(tg.tasks, task)
	tg.manager.tasksMap[task.id] = task

	tg.manager.app.addCommand("task", id).Callback(func() {
		tg.manager.run(task, nil, "command")
	})

	return task
}

//Handler sets handler to task
func (t *Task) Handler(fn func(*TaskActivity) error) *Task {
	t.handler = fn
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

//SetGroup sets group task
/*func (t *Task) SetGroup(group *TaskGroup) *Task {
	if group != nil {
		t.group = group
	}
	return t
}*/

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
