package prago

import (
	"fmt"
	"mime/multipart"
	"time"
)

//TaskActivity represents task activity
type TaskActivity struct {
	uuid      string
	task      *Task
	user      *user
	typ       string
	progress  float64
	status    string
	ended     bool
	error     error
	stoppable bool
	stopped   bool
	files     *multipart.Form
	startedAt time.Time
	endedAt   time.Time
}

//SetStatus sets progress and status for task activity
func (ta *TaskActivity) SetStatus(progress float64, status string) {
	ta.progress = progress
	ta.status = status
}

//IsStopped checks if activity is stopped
func (ta *TaskActivity) IsStopped() bool {
	ta.stoppable = true
	return ta.stopped
}

func (ta *TaskActivity) GetFile(name string) (multipart.File, error) {
	if len(ta.files.File[name]) == 0 {
		return nil, fmt.Errorf("file with id '%s' not set", name)
	}
	header := ta.files.File[name][0]
	file, err := header.Open()
	return file, err
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
