package prago

import (
	"errors"
	"net/http"
)

var (
	FileNotFoundError = errors.New("requested file is folder")
)

func MiddlewareStatic(p Request) {
	if p.IsProcessed() {
		return
	}
	w, r := p.HttpIO()
	err := ServeStatic(w, r)
	if err == nil {
		p.SetProcessed()
	}
}

func ServeStatic(w http.ResponseWriter, r *http.Request) error {

	return serveFile(w, r, http.Dir("public"), r.URL.Path)
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
		return FileNotFoundError
	}

	http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	return nil
}
