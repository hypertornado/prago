package prago

import (
	"errors"
	"net/http"
)

//ErrFileNotFound is returned when file is not found
var ErrFileNotFound = errors.New("requested file is folder")

type staticFilesHandler struct {
	paths []string
}

func newStaticHandler(paths []string) staticFilesHandler {
	return staticFilesHandler{
		paths: paths,
	}
}

func (h staticFilesHandler) serveStatic(w http.ResponseWriter, r *http.Request) bool {
	for _, v := range h.paths {
		err := serveStaticFile(w, r, http.Dir(v), r.URL.Path)
		if err == nil {
			return true
		}
	}
	return false
}

func serveStaticFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string) (err error) {
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
