package prago

import (
	"errors"
	"net/http"
)

var (
	FileNotFoundError = errors.New("requested file is folder")
	StaticDirPaths    = []string{"public"}
)

func requestMiddlewareStatic(p Request, next func()) {
	if p.IsProcessed() {
		return
	}
	if ServeStatic(p.Response(), p.Request()) {
		p.SetProcessed()
	}
	next()
}

func ServeStatic(w http.ResponseWriter, r *http.Request) bool {
	for _, v := range StaticDirPaths {
		err := serveFile(w, r, http.Dir(v), r.URL.Path)
		if err == nil {
			return true
		}
	}
	return false
}

func serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string) (err error) {
	f, err := fs.Open(name)
	if err != nil {
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return FileNotFoundError
	}

	if d.IsDir() {
		f.Close()
		f, err = fs.Open(name + "/index.html")
		if err != nil {
			return
		}

		d, err = f.Stat()
		if err != nil {
			return FileNotFoundError
		}

		if d.IsDir() {
			return FileNotFoundError
		}
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return nil
}
