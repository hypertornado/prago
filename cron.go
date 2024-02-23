package prago

import (
	"time"
)

type cronTask struct {
	id           string
	handler      func()
	lastFinished time.Time
	repeatEvery  time.Duration
}

func (app *App) initCron() {

	cronInited := time.Now()
	go func() {
		for {
			time.Sleep(1 * time.Second)
			for _, task := range app.cronTasks {
				if cronInited.Add(task.repeatEvery).Before(time.Now()) && task.lastFinished.Add(task.repeatEvery).Before(time.Now()) {
					task.handler()
					task.lastFinished = time.Now()
				}
			}
		}

	}()
}

func (app *App) addCronTask(id string, repeatEvery time.Duration, handler func()) {
	app.cronTasks = append(app.cronTasks, &cronTask{
		id:          id,
		handler:     handler,
		repeatEvery: repeatEvery,
	})

}
