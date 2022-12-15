package main

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hypertornado/prago"
)

func bindStats(app *prago.App) {

	app.MainBoard.MainDashboard.Figure(unlocalized("Počet souborů"), "sysadmin").Unit(unlocalized("souborů")).Value(func(r *prago.Request) int64 {
		var ret int64
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		filepath.Walk(homeDir+"/.pragocdn/files", func(path string, info fs.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				ret += 1
			}
			return nil
		})
		return ret
	})

	app.MainBoard.MainDashboard.Figure(unlocalized("Velikost souborů"), "sysadmin").Unit(unlocalized("bajtů")).Value(func(r *prago.Request) int64 {
		var ret int64
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		filepath.Walk(homeDir+"/.pragocdn/files", func(path string, info fs.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				ret += info.Size()
			}
			return nil
		})
		return ret
	})

}
