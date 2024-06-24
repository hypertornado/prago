package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/pragocdn/cdnclient"
)

type CDNFile struct {
	ID       int64  `prago-order-desc:"true"`
	UUID     string `prago-preview:"true"`
	Suffix   string `prago-preview:"true"`
	Checksum string `prago-preview:"true"`
	Deleted  bool   `prago-preview:"true"`

	CDNProject int64 `prago-type:"relation" prago-preview:"true"`

	Filesize int64 `prago-preview:"true"`
	Width    int64 `prago-preview:"true"`
	Height   int64 `prago-preview:"true"`

	CreatedAt time.Time
	UpdatedAt time.Time `prago-can-view:"sysadmin" prago-preview:"true"`
}

func getCDNFile(projectName, uuid string) *CDNFile {
	if !uuidRegex.MatchString(uuid) {
		return nil
	}

	project := getCDNProject(projectName)
	if project == nil {
		return nil
	}
	return prago.Query[CDNFile](app).Is("cdnproject", project.ID).Is("uuid", uuid).First()
}

func (file *CDNFile) url(size string) string {
	project := prago.Query[CDNProject](app).ID(file.CDNProject)
	if project == nil {
		panic(fmt.Errorf("can't find project id %d", file.CDNProject))
	}

	baseURL, err := app.GetSetting("base_url")
	if err != nil {
		panic(err)
	}

	account := cdnclient.NewCDNAccount(baseURL, project.Name, project.Password)

	if size == "" {
		return account.GetFileURL(file.UUID, "file."+file.Suffix)
	}

	return account.GetImageURL(file.UUID, "file."+file.Suffix, size)
}

func (file *CDNFile) Project() *CDNProject {
	return getCDNProjectFromID(file.CDNProject)
}

func (file *CDNFile) get() (*os.File, error) {
	project := prago.Query[CDNProject](app).ID(file.CDNProject)
	if project == nil {
		return nil, fmt.Errorf("can't find project id %d", file.CDNProject)
	}
	filePath := file.getDataPath()
	return os.Open(filePath)
}

func (file *CDNFile) update() {
	project := prago.Query[CDNProject](app).ID(file.CDNProject)
	if project == nil {
		panic(fmt.Errorf("can't find project id %d", file.CDNProject))
	}

	fileData, err := file.get()
	if err != nil {
		file.Deleted = true
	} else {
		defer fileData.Close()
		file.Checksum = checksum(file.getDataPath())
		metadata, err := file.getMetadata()
		if err != nil {
			panic(fmt.Sprintf("can't get metadata id %s: %s", file.UUID, err))
		}
		file.Filesize = metadata.Filesize
		file.Width = metadata.Width
		file.Height = metadata.Height
	}
	err = prago.UpdateItem(app, file)
	if err != nil {
		panic(fmt.Sprintf("can't update file id %s: %s", file.UUID, err))
	}
}

func checksum(path string) string {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("can't open file '%s' for checksum: %s", path, err))
	}
	defer f.Close()

	hasher := sha256.New()

	if _, err := io.Copy(hasher, f); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (project *CDNProject) createFile(uuid, suffix string) *CDNFile {
	file := &CDNFile{
		UUID:       uuid,
		Suffix:     suffix,
		CDNProject: project.ID,
	}

	err := prago.CreateItem(app, file)
	if err != nil {
		panic(fmt.Errorf("can't save file %s: %s", file.UUID, err))
	}
	return file
}

func (file *CDNFile) validateChecksum() {
	res := checksum(file.getDataPath())
	if res != file.Checksum {
		panic(fmt.Errorf("error while validatin checksum file %s: expecting '%s', got '%s'", file.UUID, file.Checksum, res))
	}
}

func (file *CDNFile) tempFilePath() string {
	dir := os.TempDir()
	fileName := fmt.Sprintf("pragocdn-%s.file", file.UUID)
	return path.Join(dir, fileName)
}

func bindCDNFiles(app *prago.App) {
	fileResource := prago.NewResource[CDNFile](app)
	fileResource.Name(unlocalized("CDN Soubor"), unlocalized("CDN Soubory"))

	prago.PreviewURLFunction(app, func(file *CDNFile) string {
		return file.url("")
	})

	prago.ResourceFormItemAction(app, "previewer",
		func(cdnFile *CDNFile, form *prago.Form, request *prago.Request) {
			form.AddTextInput("size", "Size")
			form.AutosubmitFirstTime = true
			form.AddSubmit("Zobrazit")
		},
		func(cdnFile *CDNFile, vc prago.ValidationContext) {
			vc.Validation().AfterContent = template.HTML(fmt.Sprintf("<img src=\"%s\">", cdnFile.url(vc.GetValue("size"))))
		},
	).Name(unlocalized("Previews"))

	filesDashboard := app.MainBoard.Dashboard(unlocalized("Soubory"))

	filesDashboard.AddTask(unlocalized("Create files form import"), "sysadmin", func(ta *prago.TaskActivity) error {
		_, err := app.GetDB().Exec("DELETE FROM cdnfile;")
		if err != nil {
			return err
		}

		projects := prago.Query[CDNProject](app).List()
		for _, project := range projects {
			filepath.Walk(cdnDirPath()+"/files/"+project.Name, func(path string, info fs.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					before, after, ok := strings.Cut(info.Name(), ".")
					if !ok {
						panic("wrong filename: " + info.Name())
					}
					project.createFile(before, after)
				}
				return nil
			})

		}
		return nil
	})

	filesDashboard.AddTask(unlocalized("Reimport files data"), "sysadmin", func(ta *prago.TaskActivity) error {
		files := prago.Query[CDNFile](app).List()
		totalLen := len(files)
		for k, file := range files {
			ta.Progress(int64(k), int64(totalLen))
			ta.Description(file.UUID)
			file.update()
		}
		return nil
	})

	filesDashboard.AddTask(unlocalized("Validate checksums"), "sysadmin", func(ta *prago.TaskActivity) error {
		files := prago.Query[CDNFile](app).List()
		totalLen := len(files)
		for k, file := range files {
			ta.Progress(int64(k), int64(totalLen))
			ta.Description(file.UUID)
			file.validateChecksum()
		}
		return nil
	})

	filesDashboard.AddTask(unlocalized("Delete thumbs cache"), "sysadmin", func(ta *prago.TaskActivity) error {
		cachePath := cdnDirPath() + "/cache"
		return os.RemoveAll(cachePath)
	})
}
