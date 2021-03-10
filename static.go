package prago

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

var errFileNotFound = errors.New("requested file is folder")

//go:embed public
var staticAdminFS embed.FS

func (app *App) initStaticFilesHandler() {
	app.staticFiles = staticFiles{}
	app.AddStaticFiles(staticAdminFS, "public")
}

//AddStaticFiles add filesystem of public files and publish them in server's root
func (app *App) AddStaticFiles(f fs.FS, pathPrefix string) {
	app.staticFiles.filesystems = append(app.staticFiles.filesystems, staticFS{
		fs:         f,
		pathPrefix: pathPrefix,
	})
}

//AddDevStaticFiles adds path for public files for development and publish them in server's root
func (app *App) AddDevStaticFiles(path string) {
	app.staticFiles.devFilesystems = append(app.staticFiles.devFilesystems, path)
}

type staticFiles struct {
	devFilesystems []string
	filesystems    []staticFS
}

type staticFS struct {
	fs         fs.FS
	path       string
	pathPrefix string
}

func (request Request) serveStatic() bool {
	if request.app.DevelopmentMode {
		for _, v := range request.app.staticFiles.devFilesystems {
			filesystem := os.DirFS(v)
			filePath := path.Join("", request.r.URL.Path[1:])
			err := request.serveStaticFile(filesystem, filePath)
			if err == nil {
				return true
			}
		}
	}

	for _, v := range request.app.staticFiles.filesystems {
		filePath := path.Join(v.pathPrefix, request.r.URL.Path)
		err := request.serveStaticFile(v.fs, filePath)
		if err == nil {
			return true
		}
	}
	return false
}

func (request Request) serveStaticFile(filesystem fs.FS, name string) (err error) {
	f, err := filesystem.Open(name)
	if err != nil {
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		return errFileNotFound
	}

	if d.IsDir() {
		f.Close()
		f, err = filesystem.Open(name + "/index.html")
		if err != nil {
			return
		}

		d, err = f.Stat()
		if err != nil {
			return errFileNotFound
		}

		if d.IsDir() {
			return errFileNotFound
		}
	}

	b, _ := ioutil.ReadAll(f)
	reader := bytes.NewReader(b)

	http.ServeContent(request.w, request.r, d.Name(), d.ModTime(), reader)
	return nil
}
