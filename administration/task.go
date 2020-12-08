package administration

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/utils"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type taskManager struct {
	admin         *Administration
	tasks         []*Task
	tasksMap      map[string]*Task
	activities    map[string]*TaskActivity
	activityMutex *sync.RWMutex
	startedAt     time.Time
}

func newTaskManager(admin *Administration) *taskManager {
	tm := &taskManager{
		admin:         admin,
		tasksMap:      make(map[string]*Task),
		activities:    make(map[string]*TaskActivity),
		activityMutex: &sync.RWMutex{},
		startedAt:     time.Now(),
	}

	return tm
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
		if v.ended && v.endedAt.Add(1*time.Hour).Before(time.Now()) {
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
	//go tm.oldTasksRemover()
	go tm.startCRON()

	tm.admin.AdminController.Get(tm.admin.GetURL("_tasks"), func(request prago.Request) {
		user := GetUser(request)
		request.SetData("tasks", tm.getTasks(user))
		request.SetData("taskmonitor", tm.getTaskMonitor(user))
		request.SetData("admin_yield", "admin_tasks")
		request.SetData("admin_title", messages.Messages.Get(user.Locale, "tasks"))
		request.RenderView("admin_layout")
	})

	tm.admin.AdminController.Get(tm.admin.GetURL("_tasks/running"), func(request prago.Request) {
		request.SetData("taskmonitor", tm.getTaskMonitor(GetUser(request)))
		request.RenderView("taskmonitor")
	})

	tm.admin.AdminController.Post(tm.admin.GetURL("_tasks/runtask"), func(request prago.Request) {
		id := request.Request().FormValue("id")
		csrf := request.Request().FormValue("csrf")
		user := GetUser(request)

		expectedToken := request.GetData("_csrfToken").(string)
		if expectedToken != csrf {
			panic("wrong token")
		}

		must(tm.runTask(id, user))
		request.Redirect(tm.admin.GetURL("_tasks"))
	})

	tm.admin.NewTask("example").SetHandler(func(t *TaskActivity) {
		var progress float64
		for {
			time.Sleep(1 * time.Second)
			t.SetStatus(progress, "example status")
			progress += 0.2
			if progress >= 1 {
				return
			}
			fmt.Println(progress)
		}
	})
}

func (tm *taskManager) runTask(id string, user User) error {
	task, ok := tm.tasksMap[id]
	if !ok {
		return fmt.Errorf("Can't find task %s", id)
	}

	if tm.admin.Authorize(user, task.permission) {
		tm.run(task, &user, "button")
		return nil
	} else {
		return fmt.Errorf("User is not authorized to run this task")
	}
}

func (tm *taskManager) getTasks(user User) (ret []TaskView) {
	for _, v := range tm.tasks {
		if tm.admin.Authorize(user, v.permission) {
			ret = append(ret, v.taskView())
		}
	}

	sort.SliceStable(ret, func(i, j int) bool {
		if collate.New(language.Czech).CompareString(ret[i].ID, ret[j].ID) <= 0 {
			return true
		}
		return false
	})

	return
}

type Task struct {
	id          string
	permission  Permission
	handler     func(*TaskActivity)
	cron        time.Duration
	lastStarted time.Time
	manager     *taskManager
}

func (t *Task) taskView() TaskView {
	return TaskView{
		ID:   t.id,
		Name: t.id,
	}
}

type TaskView struct {
	ID   string
	Name string
}

func (admin *Administration) NewTask(id string) *Task {
	_, ok := admin.taskManager.tasksMap[id]
	if ok {
		panic(fmt.Sprintf("Task '%s' already added.", id))
	}

	task := &Task{
		id:      id,
		manager: admin.taskManager,
	}

	task.manager.tasks = append(admin.taskManager.tasks, task)
	task.manager.tasksMap[task.id] = task

	return task
}

func (t *Task) SetHandler(fn func(*TaskActivity)) *Task {
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
		}()
		if t.handler != nil {
			t.handler(activity)
		}
	}()

	return activity
}

type TaskActivity struct {
	uuid      string
	task      *Task
	user      *User
	typ       string
	progress  float64
	status    string
	ended     bool
	startedAt time.Time
	endedAt   time.Time
}

func (ta *TaskActivity) SetStatus(progress float64, status string) {
	ta.progress = progress
	ta.status = status
}

type TaskActivityView struct {
	UUID                string
	TaskName            string
	Status              string
	IsDone              bool
	Progress            string
	ProgressDescription string
	StartedAt           time.Time
	StartedStr          string
	EndedStr            string
}

func (tm *taskManager) addActivity(activity *TaskActivity) {
	tm.activityMutex.Lock()
	defer tm.activityMutex.Unlock()
	tm.activities[activity.uuid] = activity
}

type TaskMonitor struct {
	Items []TaskActivityView
}

func (tm *taskManager) getTaskMonitor(user User) (ret *TaskMonitor) {
	tm.activityMutex.RLock()
	defer tm.activityMutex.RUnlock()

	ret = &TaskMonitor{}

	for _, v := range tm.activities {
		if v.user == nil {
			continue
		}
		if v.user.ID == user.ID {
			format := "15:04:05"
			startedStr := v.startedAt.Format(format)
			var endedStr string
			if v.ended {
				endedStr = v.endedAt.Format(format)
			}
			ret.Items = append(ret.Items, TaskActivityView{
				UUID:                v.uuid,
				TaskName:            v.task.id,
				Status:              v.status,
				IsDone:              v.ended,
				Progress:            fmt.Sprintf("%v", v.progress*100),
				ProgressDescription: taskProgressHuman(v.progress),
				StartedAt:           v.startedAt,
				StartedStr:          startedStr,
				EndedStr:            endedStr,
			})
		}
	}

	sort.SliceStable(ret.Items, func(i, j int) bool {
		if ret.Items[i].StartedAt.Before(ret.Items[j].StartedAt) {
			return false
		}
		return true
	})

	return
}

func taskProgressHuman(in float64) string {
	if in <= 0 {
		return ""
	}
	if in > 1 {
		return ""
	}
	return fmt.Sprintf("%.2f %%", in*100)
}
