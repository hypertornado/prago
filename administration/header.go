package administration

import (
	"github.com/hypertornado/prago"
)

type adminHeaderData struct {
	Name        string
	Language    string
	Logo        string
	UrlPrefix   string
	HomepageUrl string
	HasSearch   bool
	Items       []adminHeaderItem
}

type adminHeaderItem struct {
	Name string
	ID   string
	Url  string
}

func (admin *Administration) getHeaderData(request prago.Request) (headerData *adminHeaderData) {
	user := GetUser(request)

	var hasSearch bool
	if admin.search != nil {
		hasSearch = true
	}

	headerData = &adminHeaderData{
		Name:        admin.HumanName,
		Language:    user.Locale,
		Logo:        admin.Logo,
		UrlPrefix:   admin.prefix,
		HomepageUrl: request.App().Config.GetStringWithFallback("baseUrl", request.Request().Host),
		HasSearch:   hasSearch,
		Items:       []adminHeaderItem{},
	}

	for _, resource := range admin.getSortedResources(user.Locale) {
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
