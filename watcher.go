package prago

import (
	"os"
	"path"
	"time"
)

func (app *App) watchPath(prefix string, path string, handler func()) {
	app.Log().Printf("[%s] watching path '%s'", prefix, path)
	go func() {
		var t time.Time
		for {
			t = time.Now()
			time.Sleep(300 * time.Millisecond)
			if dirChangedAfter(path, t) {
				startTime := time.Now()
				app.Log().Printf("[%s] Compiling after change in path %s", prefix, path)
				handler()
				app.Log().Printf("[%s] ✅ Done compiling path '%s' %v", prefix, path, time.Since(startTime))
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
