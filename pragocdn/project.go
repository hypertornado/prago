package main

import (
	"time"

	"github.com/hypertornado/prago"
)

var projectResource *prago.Resource

type CDNProject struct {
	ID        int64
	Name      string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time `prago-can-view:"sysadmin" prago-preview:"true"`
}

func initCDNProjectResource() {
	projectResource = prago.NewResource[CDNProject](app)
	projectResource.Name(unlocalized("Projekt"), unlocalized("Projekty"))
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
