package prago

import (
	"errors"
	"net/http"
)

//ErrFileNotFound is returned when file is not found
var ErrFileNotFound = errors.New("requested file is folder")

type middlewareStatic struct {
	staticDirPaths []string
}

func (ms middlewareStatic) Init(app *App) error {
	ms.staticDirPaths = []string{"public"}
	paths, err := app.Config.Get("staticPaths")
	if err == nil {
		newPaths := []string{}
		for _, p := range paths.([]interface{}) {
			newPaths = append(newPaths, p.(string))
		}
		ms.staticDirPaths = newPaths
	}
	app.requestMiddlewares = append(app.requestMiddlewares, ms.requestMiddlewareStatic)
	return nil
}

func (ms middlewareStatic) requestMiddlewareStatic(p Request, next func()) {
	if p.IsProcessed() {
		return
	}
	if ms.serveStatic(p.Response(), p.Request()) {
		p.SetProcessed()
	}
	next()
}

func (ms middlewareStatic) serveStatic(w http.ResponseWriter, r *http.Request) bool {
	for _, v := range ms.staticDirPaths {
		err := ms.serveFile(w, r, http.Dir(v), r.URL.Path)
		if err == nil {
			return true
		}
	}
	return false
}

func (ms middlewareStatic) serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string) (err error) {
	f, err := fs.Open(name)
	if err != nil {
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return ErrFileNotFound
	}

	if d.IsDir() {
		f.Close()
		f, err = fs.Open(name + "/index.html")
		if err != nil {
			return
		}

		d, err = f.Stat()
		if err != nil {
			return ErrFileNotFound
		}

		if d.IsDir() {
			return ErrFileNotFound
		}
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return nil
}
