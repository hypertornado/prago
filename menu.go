package prago

import (
	"sort"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type menu struct {
	Language    string
	SearchQuery string
	Sections    []menuSection
}

type menuSection struct {
	Name  string
	Items []menuItem
}

type menuItem struct {
	Icon     string
	Name     string
	Subname  string
	URL      string
	Selected bool
	Subitems []menuItem
}

func (menu menu) GetTitle() string {
	for _, v := range menu.Sections {
		for _, v2 := range v.Items {
			if v2.Selected {
				return v2.Name
			}
		}
	}
	return ""
}

func (item menuItem) IsBoard() bool {
	if len(item.Subitems) > 0 {
		return true
	}
	return false
}

func (app *App) getMenu(request *Request) (ret menu) {
	user := request.user

	resourceSection := menuSection{
		Items: app.MainBoard.getMenuItems(request),
	}

	ret.Sections = append(ret.Sections, resourceSection)
	ret.Sections = append(ret.Sections, *getMenuUserSection(request))

	ret.Language = user.Locale
	return ret
}

func getMenuUserSection(request *Request) *menuSection {
	user := request.user
	app := request.app

	userName := user.Name
	if userName == "" {
		userName = user.Email
	}
	userSection := menuSection{
		Name:  userName,
		Items: []menuItem{},
	}
	for _, v := range app.rootActions {
		if v.method != "GET" {
			continue
		}
		if !v.isUserMenu {
			continue
		}
		if v.isHiddenInMenu {
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

		userSection.Items = append(userSection.Items, menuItem{
			Icon:     v.icon,
			Name:     v.name(user.Locale),
			URL:      fullURL,
			Selected: selected,
		})
	}

	return &userSection
}

func (board *Board) getMenuItems(request *Request) []menuItem {

	app := board.app
	var ret []menuItem
	resources := app.resources

	for _, resourceData := range resources {
		if resourceData.board != board {
			continue
		}

		if app.authorize(request.user, resourceData.canView) {
			resourceURL := resourceData.getURL("")
			var selected bool
			if request.Request().URL.Path == resourceURL {
				selected = true
			}
			if strings.HasPrefix(request.Request().URL.Path, resourceURL+"/") {
				selected = true
			}

			ret = append(ret, menuItem{
				Icon:     resourceData.icon,
				Name:     resourceData.pluralName(request.user.Locale),
				Subname:  humanizeNumber(resourceData.getCachedCount()),
				URL:      resourceURL,
				Selected: selected,
			})
		}
	}

	for _, v := range app.rootActions {
		if v.parentBoard != board {
			continue
		}
		if v.method != "GET" {
			continue
		}
		if v.isUserMenu {
			continue
		}
		if v.isHiddenInMenu {
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

		menuItem := menuItem{
			Icon:     v.icon,
			Name:     v.name(request.user.Locale),
			URL:      fullURL,
			Selected: selected,
		}

		if v.isPartOfBoard != nil && v.isPartOfBoard != app.MainBoard {
			menuItem.Subitems = v.isPartOfBoard.getMenuItems(request)
		}

		ret = append(ret, menuItem)

	}

	sortSection(ret, request.user.Locale)

	return ret
}

/*func (app *App) getSortedResources(locale string) (ret []*resourceData) {
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
}*/

func sortSection(items []menuItem, locale string) {
	collator := collate.New(language.Czech)

	sort.SliceStable(items, func(i, j int) bool {
		a := items[i]
		b := items[j]

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
