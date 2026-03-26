package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hypertornado/prago"
)

var projectResource *prago.Resource

type CDNProject struct {
	ID       int64
	Name     string
	Password string

	CDNEndpointURL string
	CDNAccessKey   string
	CDNSecretKey   string
	CDNRegion      string

	CreatedAt time.Time
	UpdatedAt time.Time `prago-can-view:"sysadmin"`
}

func initCDNProjectResource() {
	projectResource = prago.NewResource[CDNProject](app)
	projectResource.Name(unlocalized("Projekt"), unlocalized("Projekty"))

	prago.ActionForm(app, "upload-file", func(form *prago.Form, request *prago.Request) {
		projects := prago.Query[CDNProject](app).List()
		var values [][2]string
		values = append(values, [2]string{"", ""})
		for _, p := range projects {
			values = append(values, [2]string{fmt.Sprintf("%d", p.ID), p.Name})
		}
		form.AddSelect("project", "Projekt", values)
		form.AddFileInput("file", "Soubor")
		form.AddSubmit("Nahrát soubor")
	}, func(fv prago.FormValidation, request *prago.Request) {
		projectID := request.Param("project")
		if projectID == "" {
			fv.AddItemError("project", "Vyberte projekt")
		}

		multipartFiles := request.Request().MultipartForm.File["file"]
		if len(multipartFiles) != 1 {
			fv.AddItemError("file", "Vyberte soubor")
		}

		if !fv.Valid() {
			return
		}

		project := prago.Query[CDNProject](app).Is("id", projectID).First()
		if project == nil {
			fv.AddItemError("project", "Projekt nenalezen")
			return
		}

		fileHeader := multipartFiles[0]
		ext := filepath.Ext(fileHeader.Filename)
		if ext == "" {
			fv.AddItemError("file", "Soubor musí mít příponu")
			return
		}
		extension := normalizeExtension(ext[1:])

		openedFile, err := fileHeader.Open()
		if err != nil {
			fv.AddError(fmt.Sprintf("Chyba při otevírání souboru: %s", err))
			return
		}
		defer openedFile.Close()

		_, err = project.uploadFile(extension, openedFile)
		if err != nil {
			fv.AddError(fmt.Sprintf("Chyba při nahrávání souboru: %s", err))
			return
		}

		fv.Redirect("/admin/cdnfile")
	}).Name(unlocalized("Nahrát soubor")).Permission("sysadmin")
}

func getCDNProjectsMap() map[string]*CDNProject {
	var accounts = map[string]*CDNProject{}
	projects := prago.Query[CDNProject](app).List()
	for _, v := range projects {
		accounts[v.Name] = v
	}
	return accounts
}

func getCDNProjectsIDMap() map[int64]*CDNProject {
	var accounts = map[int64]*CDNProject{}
	projects := prago.Query[CDNProject](app).List()
	for _, v := range projects {
		accounts[v.ID] = v
	}
	return accounts
}

func getCDNProject(name string) *CDNProject {
	projects := <-prago.Cached(app, "get_projects_name", func() map[string]*CDNProject {
		return getCDNProjectsMap()
	})
	return projects[name]
}

func getCDNProjectFromID(id int64) *CDNProject {
	projects := <-prago.Cached(app, "get_projects_id", func() map[int64]*CDNProject {
		return getCDNProjectsIDMap()
	})
	return projects[id]
}
