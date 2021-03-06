package prago

import (
	"fmt"
	"net/url"
	"sort"
	"time"

	"github.com/hypertornado/prago/messages"
)

type taskActivityView struct {
	UUID                string
	TaskName            string
	Status              string
	IsDone              bool
	IsStopped           bool
	IsError             bool
	Progress            string
	ProgressDescription string
	StartedAt           time.Time
	StartedStr          string
	EndedStr            string
	Actions             []taskActivityViewAction
}

type taskActivityViewAction struct {
	Name string
	URL  string
}

func (tm *taskManager) getTaskMonitor(user User) (ret *taskMonitor) {
	tm.activityMutex.RLock()
	defer tm.activityMutex.RUnlock()

	ret = &taskMonitor{}

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

			var actions []taskActivityViewAction
			if !v.ended && v.stoppable && !v.stopped {
				var u url.Values = map[string][]string{}
				u.Add("uuid", v.uuid)
				randomness := tm.app.Config.GetString("random")
				u.Add("csrf", user.CSRFToken(randomness))
				actions = append(actions, taskActivityViewAction{"◼", "_tasks/stoptask?" + u.Encode()})
			}

			if v.ended {
				var u url.Values = map[string][]string{}
				u.Add("uuid", v.uuid)
				randomness := tm.app.Config.GetString("random")
				u.Add("csrf", user.CSRFToken(randomness))
				actions = append(actions, taskActivityViewAction{"✘", "_tasks/deletetask?" + u.Encode()})
			}

			status := v.status

			var isError bool
			if v.error != nil {
				isError = true
				status = v.error.Error()
			}

			ret.Items = append(ret.Items, taskActivityView{
				UUID:                v.uuid,
				TaskName:            v.task.id,
				Status:              status,
				IsDone:              v.ended,
				IsStopped:           v.stopped,
				IsError:             isError,
				Progress:            fmt.Sprintf("%v", v.progress*100),
				ProgressDescription: taskProgressHuman(v.progress),
				StartedAt:           v.startedAt,
				StartedStr:          startedStr,
				EndedStr:            endedStr,
				Actions:             actions,
			})
		}
	}

	sort.SliceStable(ret.Items, func(i, j int) bool {
		if ret.Items[i].StartedAt.Before(ret.Items[j].StartedAt) {
			return false
		}
		return true
	})

	if len(ret.Items) > 0 {
		ret.Name = messages.Messages.Get(user.Locale, "tasks_runned")
	}

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
