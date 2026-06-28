package prago

import (
	"log"
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
					task.handle()
				}
			}
		}

	}()
}

func (ct *cronTask) handle() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("recovering from crontask handle panic: %v", err)
		}
	}()
	ct.handler()
	ct.lastFinished = time.Now()
}

func (app *App) addCronTask(id string, repeatEvery time.Duration, handler func()) {
	app.cronTasks = append(app.cronTasks, &cronTask{
		id:          id,
		handler:     handler,
		repeatEvery: repeatEvery,
	})

}
