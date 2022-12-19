package main

import (
	"context"
	"time"

	"github.com/hypertornado/prago"
)

var projectResource *prago.Resource[CDNProject]

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
	projects := projectResource.Query(context.Background()).List()
	for _, v := range projects {
		accounts[v.Name] = v
	}
	return accounts
}

func getCDNProject(id string) *CDNProject {
	projects := <-prago.Cached(app, "get_projects", func(ctx context.Context) map[string]*CDNProject {
		return getCDNProjectsMap()
	})
	return projects[id]
}
