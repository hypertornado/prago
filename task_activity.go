package prago

// TaskActivity represents task activity
type TaskActivity struct {
	task          *Task
	notification  *Notification
	stoppedByUser bool
}

func (ta *TaskActivity) Progress(finishedSoFar, total int64) {
	if ta == nil {
		return
	}
	if ta.stoppedByUser {
		panic("task already stopped by user")
	}
	progress := float64(finishedSoFar) / float64(total)
	ta.notification.SetProgress(&progress)
}

func (ta *TaskActivity) Description(description string) {
	if ta == nil {
		return
	}
	if ta.stoppedByUser {
		panic("task already stopped by user")
	}
	ta.notification.description = description
}
