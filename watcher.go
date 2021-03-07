package prago

import (
	"os"
	"path"
	"time"
)

func (app *App) watchPath(path string, handler func()) {
	go func() {
		var t time.Time
		for {
			t = time.Now()
			time.Sleep(300 * time.Millisecond)
			if dirChangedAfter(path, t) {
				handler()
			}
		}
	}()
}

func dirChangedAfter(dirPath string, t time.Time) bool {
	files, err := os.ReadDir(dirPath)
	must(err)

	for _, v := range files {
		i, err := v.Info()
		must(err)
		if t.Before(i.ModTime()) {
			return true
		}

		newPath := path.Join(dirPath, v.Name())
		if v.IsDir() {
			c := dirChangedAfter(newPath, t)
			if c {
				return true
			}
		}
	}

	return false
}
