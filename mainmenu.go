package prago

import (
	"strings"

	"github.com/hypertornado/prago/utils"
)

type mainMenu struct {
	Logo             string
	Language         string
	URLPrefix        string
	AdminHomepageURL string
	SearchQuery      string
	Sections         []mainMenuSection
	HasSearch        bool
}

type mainMenuSection struct {
	Name  string
	Items []mainMenuItem
}

type mainMenuItem struct {
	Name     string
	Subname  string
	URL      string
	Selected bool
}

func (menu mainMenu) GetTitle() string {
	for _, v := range menu.Sections {
		for _, v2 := range v.Items {
			if v2.Selected {
				return v2.Name
			}
		}
	}
	return ""
}

func (app *App) getMainMenu(request Request) (ret mainMenu) {
	user := request.GetUser()

	adminSectionName := app.name(user.Locale)
	if app.logo != "" {
		adminSectionName = ""
	}
	adminSection := mainMenuSection{
		Name: adminSectionName,
	}

	for _, v := range app.rootActions {
		if v.method != "GET" {
			continue
		}
		if v.isUserMenu {
			continue
		}
		if v.isHiddenMenu {
			continue
		}

		var selected bool
		fullURL := app.GetAdminURL(v.url)
		if request.Request().URL.Path == fullURL {
			selected = true
		}

		adminSection.Items = append(adminSection.Items, mainMenuItem{
			Name:     v.name(user.Locale),
			URL:      fullURL,
			Selected: selected,
		})
	}

	ret.Sections = append(ret.Sections, adminSection)

	resourceSection := mainMenuSection{
		Name: messages.Get(user.Locale, "admin_tables"),
	}
	for _, resource := range app.getSortedResources(user.Locale) {
		if app.Authorize(user, resource.canView) {
			resourceURL := resource.getURL("")
			var selected bool
			if request.Request().URL.Path == resourceURL {
				selected = true
			}
			if strings.HasPrefix(request.Request().URL.Path, resourceURL+"/") {
				selected = true
			}

			resourceSection.Items = append(resourceSection.Items, mainMenuItem{
				Name:     resource.name(user.Locale),
				Subname:  utils.HumanizeNumber(resource.getCachedCount()),
				URL:      resourceURL,
				Selected: selected,
			})
		}
	}
	ret.Sections = append(ret.Sections, resourceSection)

	userName := user.Name
	if userName == "" {
		userName = user.Email
	}
	randomness := app.ConfigurationGetString("random")
	userSection := mainMenuSection{
		Name:  userName,
		Items: []mainMenuItem{},
	}
	for _, v := range app.rootActions {
		if v.method != "GET" {
			continue
		}
		if !v.isUserMenu {
			continue
		}
		if v.isHiddenMenu {
			continue
		}

		var selected bool
		fullURL := app.GetAdminURL(v.url)
		if request.Request().URL.Path == fullURL {
			selected = true
		}

		if v.url == "logout" {
			fullURL += "?_csrfToken=" + user.csrfToken(randomness)
		}

		userSection.Items = append(userSection.Items, mainMenuItem{
			Name:     v.name(user.Locale),
			URL:      fullURL,
			Selected: selected,
		})
	}
	ret.Sections = append(ret.Sections, userSection)

	ret.Logo = app.logo
	ret.URLPrefix = adminPathPrefix
	ret.Language = user.Locale
	ret.AdminHomepageURL = app.GetAdminURL("")

	if app.search != nil {
		ret.HasSearch = true
	}

	return ret
}
