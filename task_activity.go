package prago

// TaskActivity represents task activity
type TaskActivity struct {
	task          *Task
	notification  *Notification
	stoppedByUser bool
}

// SetStatus sets progress and status for task activity
func (ta *TaskActivity) SetStatus(progress float64, status string) {
	if ta.stoppedByUser {
		panic("task already stopped by user")
	}
	ta.notification.description = status
	ta.notification.SetProgress(&progress)
}
