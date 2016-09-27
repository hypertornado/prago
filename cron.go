package prago

import (
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
				task.task()
				now := time.Now()
				task.lastExecuted = now
				task.scheduled = task.timer(now)
			}
			c.mutex.Unlock()
		}
	}
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
func (a *App) AddCronTask(name string, task func(), timer func(time.Time) time.Time) {
	a.cron.mutex.Lock()
	ct := &cronTask{
		name:      name,
		task:      task,
		timer:     timer,
		scheduled: timer(time.Now()),
	}
	a.cron.tasks[name] = ct
	a.cron.mutex.Unlock()
}
