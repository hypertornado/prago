package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hypertornado/prago"
)

func bindStats(app *prago.App) {

	app.MainBoard.MainDashboard.Figure(unlocalized("Počet souborů"), "sysadmin").Unit(unlocalized("souborů")).Value(func(r *prago.Request) int64 {
		var ret int64
		filepath.Walk(cdnDirPath()+"/files", func(path string, info fs.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				ret += 1
			}
			return nil
		})
		return ret
	})

	app.MainBoard.MainDashboard.Figure(unlocalized("Velikost souborů"), "sysadmin").Unit(unlocalized("bajtů")).Value(func(r *prago.Request) int64 {
		var ret int64
		filepath.Walk(cdnDirPath()+"/files", func(path string, info fs.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				ret += info.Size()
			}
			return nil
		})
		return ret
	})

	app.MainBoard.Dashboard(unlocalized("Statistiky souborů")).Table(func(r *prago.Request) *prago.Table {
		ret := app.Table()

		projects := prago.GetResource[CDNProject](app).Query(r.Request().Context()).Order("Name").List()

		ret.Header("project", "size name", "count", "size in bytes")

		for _, project := range projects {
			count, size := getOriginalsStats(project.Name)
			ret.Row(project.Name, "original", count, size)

			sizes := getSizes(project.Name)
			for _, sizeName := range sizes {
				count, size := getCacheStats(project.Name, sizeName)
				ret.Row("", sizeName, count, size)
			}

			//ret.Row("", "original", project.Name)
		}
		return ret
	}, "sysadmin")
}

func getOriginalsStats(projectName string) (count, size int64) {
	filepath.Walk(cdnDirPath()+"/files/"+projectName, func(path string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count += 1
			size += info.Size()
		}
		return nil
	})
	return
}

func getCacheStats(projectName, sizeName string) (count, size int64) {
	filepath.Walk(cdnDirPath()+"/cache/"+projectName+"/"+sizeName, func(path string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count += 1
			size += info.Size()
		}
		return nil
	})
	return
}

func getSizes(projectName string) (ret []string) {
	files, err := os.ReadDir(cdnDirPath() + "/cache/" + projectName)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() && !strings.HasPrefix(".", file.Name()) {
			ret = append(ret, file.Name())
		}
	}
	return
}
