package prago

import (
	"context"
	"sort"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type menu struct {
	Language    string
	SearchQuery string
	Sections    []*menuSection
}

type menuSection struct {
	Name  string
	Items []menuItem
}

type menuItem struct {
	Icon        string
	Name        string
	URL         string
	Selected    bool
	Subitems    []menuItem
	Expanded    bool
	IsBoard     bool
	IsMainBoard bool
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

func (app *App) initMenuAPI() {

	app.API("resource-counts").Permission(loggedPermission).HandlerJSON(func(request *Request) any {
		return getResourceCountsMap(request)
	})

}

func (app *App) getMenu(userData UserData, urlPath, csrfToken string) (ret menu) {
	items, _ := app.MainBoard.getMainItems(userData, urlPath, csrfToken)

	resourceSection := &menuSection{
		Items: items,
	}

	ret.Sections = append(ret.Sections, resourceSection)
	ret.Sections = append(ret.Sections, app.getMenuUserSection(userData, urlPath, csrfToken))

	ret.Language = userData.Locale()
	return ret
}

func getResourceCountsMap(request *Request) map[string]string {
	app := request.app
	ret := make(map[string]string)

	for _, v := range app.resources {
		if request.Authorize(v.canView) {
			url := v.getURL("")
			count := v.getCachedCount(context.TODO())
			ret[url] = humanizeNumber(count)
		}

	}
	return ret
}

func (app *App) getMenuUserSection(userData UserData, urlPath, csrfToken string) *menuSection {
	userName := userData.Name()
	mainItems, _ := app.MainBoard.getItems(userData, urlPath, true, csrfToken)
	userSection := menuSection{
		Name:  userName,
		Items: mainItems,
	}

	return &userSection
}

//request.Request().URL.Path

func (board *Board) getMainItems(userData UserData, urlPath string, csrfToken string) ([]menuItem, bool) {
	return board.getItems(userData, urlPath, false, csrfToken)
}

func (board *Board) getItems(userData UserData, urlPath string, isUserMenu bool, csrfToken string) ([]menuItem, bool) {
	app := board.app
	var ret []menuItem

	var isExpanded bool

	if !isUserMenu {
		resources := app.resources
		for _, resourceData := range resources {
			if resourceData.board != board {
				continue
			}

			if userData.Authorize(resourceData.canView) {
				resourceURL := resourceData.getURL("")
				var selected bool
				if urlPath == resourceURL {
					selected = true
				}
				if strings.HasPrefix(urlPath, resourceURL+"/") {
					selected = true
				}

				if selected {
					isExpanded = true
				}

				ret = append(ret, menuItem{
					Icon: resourceData.icon,
					Name: resourceData.pluralName(userData.Locale()),
					//Subname:  humanizeNumber(resourceData.getCachedCount(request.r.Context())),
					URL:      resourceURL,
					Selected: selected,
				})
			}
		}
	}

	for _, v := range app.rootActions {
		if v.parentBoard != board {
			continue
		}
		if v.method != "GET" {
			continue
		}
		if v.isUserMenu != isUserMenu {
			continue
		}
		if v.isHiddenInMenu {
			continue
		}
		if !userData.Authorize(v.permission) {
			continue
		}

		var selected bool
		fullURL := app.getAdminURL(v.url)
		if urlPath == fullURL {
			selected = true
			isExpanded = true
		}

		if fullURL == "/admin/logout" {
			fullURL += "?_csrfToken=" + csrfToken
		}

		var isBoard, isMainBoard bool
		if v.isPartOfBoard != nil {
			if v.isPartOfBoard.isEmpty(userData, urlPath) {
				continue
			}

			isBoard = true
			if v.isPartOfBoard.IsMainBoard() {
				isMainBoard = true
			}
		}

		menuItem := menuItem{
			Icon:        v.icon,
			Name:        v.name(userData.Locale()),
			URL:         fullURL,
			Selected:    selected,
			Expanded:    selected,
			IsBoard:     isBoard,
			IsMainBoard: isMainBoard,
		}

		if v.isPartOfBoard != nil && v.isPartOfBoard != app.MainBoard {
			subitems, subitemsIsExpanded := v.isPartOfBoard.getMainItems(userData, urlPath, csrfToken)
			if subitemsIsExpanded {
				menuItem.Expanded = true
				isExpanded = true
			}
			menuItem.Subitems = subitems
		}

		ret = append(ret, menuItem)

	}

	sortSection(ret, userData.Locale())

	return ret, isExpanded
}

func sortSection(items []menuItem, locale string) {
	collator := collate.New(language.Czech)

	sort.SliceStable(items, func(i, j int) bool {
		a := items[i]
		b := items[j]

		if a.IsMainBoard {
			return true
		}

		if b.IsMainBoard {
			return false
		}

		if a.IsBoard && !b.IsBoard {
			return true
		}

		if !a.IsBoard && b.IsBoard {
			return false
		}

		if collator.CompareString(a.Name, b.Name) <= 0 {
			return true
		} else {
			return false
		}
	})
}
