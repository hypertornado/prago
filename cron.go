package prago

import (
	"fmt"
	"sync"
	"time"
)

type cron struct {
	mutex sync.Mutex
	tasks map[string]*cronTask
}

func (c *cron) scheduler() {
	for {
		//TODO: dont wait fixed amount of time
		time.Sleep(1 * time.Second)
		for _, task := range c.tasks {
			c.mutex.Lock()
			if task.scheduled.Before(time.Now()) {
				runCronTask(task.task)
				now := time.Now()
				task.lastExecuted = now
				task.scheduled = task.timer(now)
			}
			c.mutex.Unlock()
		}
	}
}

func runCronTask(task func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in runCronTask", r)
		}
	}()
	task()
}

type cronTask struct {
	name         string
	task         func()
	timer        func(time.Time) time.Time
	lastExecuted time.Time
	scheduled    time.Time
}

func newCron() *cron {
	tasks := make(map[string]*cronTask)
	cr := &cron{
		tasks: tasks,
	}
	go cr.scheduler()
	return cr
}

//AddCronTask ads task function with name, which is executed regularly
//timer functions returns next execution time
func (app *App) AddCronTask(name string, task func(), timer func(time.Time) time.Time) {
	app.cron.mutex.Lock()
	ct := &cronTask{
		name:      name,
		task:      task,
		timer:     timer,
		scheduled: timer(time.Now()),
	}
	app.cron.tasks[name] = ct
	app.cron.mutex.Unlock()
}
