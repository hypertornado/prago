package prago

import (
	"sort"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type mainMenu struct {
	HasLogo          bool
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
	Icon     string
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

func (app *App) getMainMenu(request *Request) (ret mainMenu) {
	user := request.user
	if app.logo != nil {
		ret.HasLogo = true
	}

	adminSection := mainMenuSection{}

	for _, v := range app.rootActions {
		if v.method != "GET" {
			continue
		}
		if v.isUserMenu {
			continue
		}
		if v.isHiddenInMainMenu {
			continue
		}
		if !request.app.authorize(request.user, v.permission) {
			continue
		}

		var selected bool
		fullURL := app.getAdminURL(v.url)
		if request.Request().URL.Path == fullURL {
			selected = true
		}

		adminSection.Items = append(adminSection.Items, mainMenuItem{
			Icon:     v.icon,
			Name:     v.name(user.Locale),
			URL:      fullURL,
			Selected: selected,
		})
	}

	resourceSection2 := app.getResourcesMainMenuSection(request, user)

	ret.Sections = append(ret.Sections, resourceSection2)

	userName := user.Name
	if userName == "" {
		userName = user.Email
	}
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
		if v.isHiddenInMainMenu {
			continue
		}
		if !request.app.authorize(request.user, v.permission) {
			continue
		}

		var selected bool
		fullURL := app.getAdminURL(v.url)
		if request.Request().URL.Path == fullURL {
			selected = true
		}

		if v.url == "logout" {
			fullURL += "?_csrfToken=" + app.generateCSRFToken(user)
		}

		userSection.Items = append(userSection.Items, mainMenuItem{
			Name:     v.name(user.Locale),
			URL:      fullURL,
			Selected: selected,
		})
	}
	ret.Sections = append(ret.Sections, userSection)

	ret.URLPrefix = adminPathPrefix
	ret.Language = user.Locale
	ret.AdminHomepageURL = app.getAdminURL("")

	if app.search != nil {
		ret.HasSearch = true
	}

	return ret
}

func (app *App) getResourcesMainMenuSection(request *Request, user *user) mainMenuSection {
	resourceSection := mainMenuSection{
		//Name: messages.Get(user.Locale, "admin_tables"),
	}
	for _, resourceData := range app.getSortedResources(user.Locale) {
		if app.authorize(user, resourceData.canView) {
			resourceURL := resourceData.getURL("")
			var selected bool
			if request.Request().URL.Path == resourceURL {
				selected = true
			}
			if strings.HasPrefix(request.Request().URL.Path, resourceURL+"/") {
				selected = true
			}

			resourceSection.Items = append(resourceSection.Items, mainMenuItem{
				Icon:     resourceData.icon,
				Name:     resourceData.pluralName(user.Locale),
				Subname:  humanizeNumber(resourceData.getCachedCount()),
				URL:      resourceURL,
				Selected: selected,
			})
		}
	}

	for _, v := range app.rootActions {
		if v.method != "GET" {
			continue
		}
		if v.isUserMenu {
			continue
		}
		if v.isHiddenInMainMenu {
			continue
		}
		if !request.app.authorize(request.user, v.permission) {
			continue
		}

		var selected bool
		fullURL := app.getAdminURL(v.url)
		if request.Request().URL.Path == fullURL {
			selected = true
		}

		resourceSection.Items = append(resourceSection.Items, mainMenuItem{
			Icon:     v.icon,
			Name:     v.name(user.Locale),
			URL:      fullURL,
			Selected: selected,
		})
	}

	sortSection(resourceSection, user.Locale)

	return resourceSection
}

func (app *App) getSortedResources(locale string) (ret []*resourceData) {
	collator := collate.New(language.Czech)

	ret = app.resources
	sort.SliceStable(ret, func(i, j int) bool {
		a := ret[i]
		b := ret[j]

		if collator.CompareString(a.pluralName(locale), b.pluralName(locale)) <= 0 {
			return true
		} else {
			return false
		}
	})
	return
}

func sortSection(section mainMenuSection, locale string) {
	collator := collate.New(language.Czech)

	sort.SliceStable(section.Items, func(i, j int) bool {
		a := section.Items[i]
		b := section.Items[j]

		if a.URL == "/admin" {
			return true
		}

		if b.URL == "/admin" {
			return false
		}

		if collator.CompareString(a.Name, b.Name) <= 0 {
			return true
		} else {
			return false
		}
	})
}
