package administration

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

func (admin *Administration) getHeaderData(request prago.Request) (headerData *adminHeaderData) {
	user := GetUser(request)

	headerData = &adminHeaderData{
		Name:        admin.HumanName,
		Language:    user.Locale,
		Logo:        admin.Logo,
		Background:  admin.Background,
		UrlPrefix:   admin.Prefix,
		HomepageUrl: request.App().Config.GetStringWithFallback("baseUrl", request.Request().Host),
		Items:       []adminHeaderItem{},
	}

	for _, resource := range admin.Resources {
		if admin.Authorize(user, resource.CanView) {
			headerData.Items = append(headerData.Items, adminHeaderItem{
				Name: resource.HumanName(user.Locale),
				ID:   resource.ID,
				Url:  admin.GetURL(resource.ID),
			})
		}
	}
	return
}
