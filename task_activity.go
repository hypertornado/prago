package prago

import (
	"fmt"
	"io"
	"mime/multipart"
)

// TaskActivity represents task activity
type TaskActivity struct {
	task          *Task
	notification  *Notification
	stoppedByUser bool
	files         *multipart.Form
}

// SetStatus sets progress and status for task activity
func (ta *TaskActivity) SetStatus(progress float64, status string) {
	if ta.stoppedByUser {
		panic("task already stopped by user")
	}
	ta.notification.description = status
	ta.notification.SetProgress(&progress)
}

func (ta *TaskActivity) GetFile(name string) (multipart.File, error) {
	if len(ta.files.File[name]) == 0 {
		return nil, fmt.Errorf("file with id '%s' not set", name)
	}
	header := ta.files.File[name][0]
	file, err := header.Open()
	return file, err
}

func (ta *TaskActivity) GetFileContent() []byte {
	if len(ta.task.files) == 0 {
		panic("no files for task set")
	}

	file, err := ta.GetFile(ta.task.files[0].ID)
	must(err)
	content, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	return content
}
