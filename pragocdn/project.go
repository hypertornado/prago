package main

import (
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

func unlocalized(in string) func(string) string {
	return func(string) string {
		return in
	}
}

func initCDNProjectResource() {
	projectResource = prago.NewResource[CDNProject](app)
	projectResource.Name(unlocalized("Projekt"), unlocalized("Projekty"))
}
