package main

import (
	"io/fs"
	"path/filepath"

	"github.com/hypertornado/prago"
)

func bindStats(app *prago.App) {

	app.MainBoard.MainDashboard.Figure(unlocalized("Počet souborů"), "sysadmin").Unit(unlocalized("souborů")).Value(func(r *prago.Request) int64 {
		files := prago.Query[CDNFile](app).List()
		return int64(len(files))
	})

	app.MainBoard.MainDashboard.Figure(unlocalized("Velikost souborů"), "sysadmin").Unit(unlocalized("bajtů")).Value(func(r *prago.Request) int64 {
		files := prago.Query[CDNFile](app).List()
		var ret int64
		for _, file := range files {
			ret += file.Filesize
		}
		return ret
	})

	app.MainBoard.MainDashboard.Figure(unlocalized("Datová velikost souborů"), "sysadmin").Unit(unlocalized("bajtů")).Value(func(r *prago.Request) int64 {
		var ret int64
		filepath.Walk(cdnDirPath()+"/data", func(path string, info fs.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				ret += info.Size()
			}
			return nil
		})
		return ret
	})

}
