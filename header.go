package prago

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

func (app *App) getHeaderData(request Request) (headerData *adminHeaderData) {
	user := GetUser(request)

	var hasSearch bool
	if app.search != nil {
		hasSearch = true
	}

	headerData = &adminHeaderData{
		Name:        app.HumanName,
		Language:    user.Locale,
		Logo:        app.Logo,
		UrlPrefix:   app.prefix,
		HomepageUrl: app.Config.GetStringWithFallback("baseUrl", request.Request().Host),
		HasSearch:   hasSearch,
		Items:       []adminHeaderItem{},
	}

	for _, resource := range app.getSortedResources(user.Locale) {
		if app.Authorize(user, resource.CanView) {
			headerData.Items = append(headerData.Items, adminHeaderItem{
				Name: resource.HumanName(user.Locale),
				ID:   resource.ID,
				Url:  app.GetAdminURL(resource.ID),
			})
		}
	}
	return
}
