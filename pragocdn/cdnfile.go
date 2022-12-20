package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
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

	fileResource := prago.GetResource[CDNFile](app)
	return fileResource.Query(context.Background()).Is("cdnproject", project.ID).Is("uuid", uuid).First()
}

func (file *CDNFile) url() string {
	projectResource := prago.GetResource[CDNProject](app)
	project := projectResource.Query(context.Background()).ID(file.CDNProject)
	if project == nil {
		panic(fmt.Errorf("can't find project id %d", file.CDNProject))
	}

	baseURL := app.MustGetSetting(context.Background(), "base_url")

	account := cdnclient.NewCDNAccount(baseURL, project.Name, project.Password)

	return account.GetFileURL(file.UUID, "file."+file.Suffix)
}

func (file *CDNFile) Project() *CDNProject {
	return getCDNProjectFromID(file.CDNProject)
}

func (file *CDNFile) get() (*os.File, error) {
	projectResource := prago.GetResource[CDNProject](app)
	project := projectResource.Query(context.Background()).ID(file.CDNProject)
	if project == nil {
		return nil, fmt.Errorf("can't find project id %d", file.CDNProject)
	}
	filePath := file.getFilePath()
	return os.Open(filePath)
}

func (file *CDNFile) update() {
	projectResource := prago.GetResource[CDNProject](app)
	project := projectResource.Query(context.Background()).ID(file.CDNProject)
	if project == nil {
		panic(fmt.Errorf("can't find project id %d", file.CDNProject))
	}

	fileResource := prago.GetResource[CDNFile](app)
	fileData, err := file.get()
	if err != nil {
		file.Deleted = true
	} else {
		defer fileData.Close()

		file.Checksum = checksum(fileData)
		metadata, err := getMetadata(project.Name, file.UUID)
		if err != nil {
			panic(fmt.Sprintf("can't get metadata id %s: %s", file.UUID, err))
		}
		file.Filesize = metadata.Filesize
		file.Width = metadata.Width
		file.Height = metadata.Height
	}
	err = fileResource.Update(context.Background(), file)
	if err != nil {
		panic(fmt.Sprintf("can't update file id %s: %s", file.UUID, err))
	}
}

func checksum(file *os.File) string {
	hasher := sha256.New()

	if _, err := io.Copy(hasher, file); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (project *CDNProject) createFile(uuid, suffix string) *CDNFile {
	fileResource := prago.GetResource[CDNFile](app)
	file := &CDNFile{
		UUID:       uuid,
		Suffix:     suffix,
		CDNProject: project.ID,
	}

	err := fileResource.Create(context.Background(), file)
	if err != nil {
		panic(fmt.Errorf("can't save file %s: %s", file.UUID, err))
	}
	return file
}

func (file *CDNFile) copyToChecksumFormat() {

	originalPath := file.getFilePath()
	_, err := os.Stat(originalPath)
	if err != nil {
		panic(fmt.Errorf("cant open file %s: %s", file.UUID, err))
	}

	targetPath := file.getDataPath()
	_, err = os.Stat(targetPath)
	if err == nil {
		return
	}

	targetDir := file.getDataDirectoryPath()
	err = os.MkdirAll(targetDir, 0777)
	if err != nil {
		panic(err)
	}

	originalFile, err := os.Open(originalPath)
	if err != nil {
		panic(fmt.Errorf("can't open original file while copying %s: %s", file.UUID, err))
	}
	defer originalFile.Close()

	f, err := os.Create(targetPath)
	if err != nil {
		panic(fmt.Errorf("can't create file while copying %s: %s", file.UUID, err))
	}
	defer f.Close()

	_, err = io.Copy(f, originalFile)
	if err != nil {
		panic(fmt.Errorf("error while copying copying %s: %s", file.UUID, err))
	}

}

func bindCDNFiles(app *prago.App) {
	fileResource := prago.NewResource[CDNFile](app)
	fileResource.Name(unlocalized("CDN Soubor"), unlocalized("CDN Soubory"))

	fileResource.PreviewURLFunction(func(file *CDNFile) string {
		return file.url()
	})

	tg := app.TaskGroup(unlocalized("Soubory"))

	tg.Task(unlocalized("Create files form import")).Handler(func(ta *prago.TaskActivity) error {
		_, err := app.GetDB().Exec("DELETE FROM cdnfile;")
		if err != nil {
			return err
		}

		projects := prago.GetResource[CDNProject](app).Query(context.Background()).List()
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

	tg.Task(unlocalized("Reimport files data")).Handler(func(ta *prago.TaskActivity) error {
		fileResource := prago.GetResource[CDNFile](app)

		files := fileResource.Query(context.Background()).List()
		totalLen := len(files)
		for k, file := range files {
			ta.SetStatus(float64(k)/float64(totalLen), file.UUID)
			file.update()
		}
		return nil
	})

	tg.Task(unlocalized("Copy data to new checkum format")).Handler(func(ta *prago.TaskActivity) error {
		fileResource := prago.GetResource[CDNFile](app)
		files := fileResource.Query(context.Background()).List()
		totalLen := len(files)
		for k, file := range files {
			ta.SetStatus(float64(k)/float64(totalLen), file.UUID)
			file.copyToChecksumFormat()
		}
		return nil

	})
}
