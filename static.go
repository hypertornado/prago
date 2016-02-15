package prago

import (
	"errors"
	"net/http"
)

var (
	FileNotFoundError = errors.New("requested file is folder")
	StaticDirPath     = "public"
)

func requestMiddlewareStatic(p Request, next func()) {
	if p.IsProcessed() {
		return
	}
	err := ServeStatic(p.Response(), p.Request())
	if err == nil {
		p.SetProcessed()
	}

	next()
}

func ServeStatic(w http.ResponseWriter, r *http.Request) error {
	return serveFile(w, r, http.Dir(StaticDirPath), r.URL.Path)
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
