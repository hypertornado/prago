package admin

import (
	"github.com/hypertornado/prago"
)

type adminHeaderData struct {
	Name        string
	Language    string
	Logo        string
	Background  string
	UrlPrefix   string
	HomepageUrl string
	Items       []adminHeaderItem
}

type adminHeaderItem struct {
	Name string
	ID   string
	Url  string
}

func (a *Admin) getHeaderData(request prago.Request) (headerData *adminHeaderData) {

	user := GetUser(request)
	locale := GetLocale(request)

	headerData = &adminHeaderData{
		Name:        a.HumanName,
		Language:    locale,
		Logo:        a.Logo,
		Background:  a.Background,
		UrlPrefix:   a.Prefix,
		HomepageUrl: request.App().Config.GetStringWithFallback("baseUrl", request.Request().Host),
		Items:       []adminHeaderItem{},
	}

	for _, resource := range a.Resources {
		if resource.HasView && resource.Authenticate(user) {
			headerData.Items = append(headerData.Items, adminHeaderItem{
				Name: resource.Name(locale),
				ID:   resource.ID,
				Url:  a.Prefix + "/" + resource.ID,
			})
		}
	}
	return
}
