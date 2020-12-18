package administration

import (
	"time"
)

type TaskActivity struct {
	uuid      string
	task      *Task
	user      *User
	typ       string
	progress  float64
	status    string
	ended     bool
	error     error
	stoppable bool
	stopped   bool
	startedAt time.Time
	endedAt   time.Time
}

func (ta *TaskActivity) SetStatus(progress float64, status string) {
	ta.progress = progress
	ta.status = status
}

func (ta *TaskActivity) IsStopped() bool {
	ta.stoppable = true
	return ta.stopped
}

type taskMonitor struct {
	Name  string
	Items []taskActivityView
}

func (tm *taskManager) addActivity(activity *TaskActivity) {
	tm.activityMutex.Lock()
	defer tm.activityMutex.Unlock()
	tm.activities[activity.uuid] = activity
}
