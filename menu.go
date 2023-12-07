package prago

import (
	"context"
	"fmt"
	"sort"
	"strconv"
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
			ret := getTitleFromMenuSubsections(v2)
			if len(ret) > 0 {
				return strings.Join(ret, " · ")
			}
		}
	}
	return ""
}

func getTitleFromMenuSubsections(item menuItem) []string {
	if item.Selected {
		return []string{
			item.Name,
		}
	}

	for _, v := range item.Subitems {
		items := getTitleFromMenuSubsections(v)
		if len(items) > 0 {
			ret := append(items, item.Name)
			return ret
		}
	}
	return []string{}
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
	mainItems, _ := app.MainBoard.getMenuItems(userData, urlPath, true, csrfToken)
	userSection := menuSection{
		Name:  userName,
		Items: mainItems,
	}

	return &userSection
}

func (board *Board) getMainItems(userData UserData, urlPath string, csrfToken string) ([]menuItem, bool) {
	return board.getMenuItems(userData, urlPath, false, csrfToken)
}

func (board *Board) getMenuItems(userData UserData, urlPath string, isUserMenu bool, csrfToken string) ([]menuItem, bool) {
	app := board.app
	var ret []menuItem

	var isExpanded bool
	var dontSortByName bool

	if board.parentResource != nil {
		parentResource := board.parentResource
		resourceURLPath := board.parentResource.getURL("")

		var itemID int

		if strings.HasPrefix(urlPath, resourceURLPath) && len(urlPath) > len(resourceURLPath) {
			isExpanded = true

			beforeStr, _, _ := strings.Cut(urlPath[len(resourceURLPath)+1:], "/")
			if true {
				itemID, _ = strconv.Atoi(beforeStr)
			}
		}

		dontSortByName = true
		navigation := board.parentResource.getResourceNavigation(userData, "")
		for k, v := range navigation.Tabs {
			if k == 0 {
				continue
			}

			var selected bool
			if urlPath == v.URL {
				isExpanded = true
				selected = true
			}

			ret = append(ret, menuItem{
				Icon:     v.Icon,
				Name:     v.Name,
				URL:      v.URL,
				Selected: selected,
			})
		}

		if itemID > 0 {

			item := parentResource.query(context.Background()).ID(itemID)

			itemPreviewData := parentResource.previewer(userData, item)

			ret = append(ret, menuItem{
				Icon:     iconView,
				Name:     itemPreviewData.Name(),
				URL:      parentResource.getURL(fmt.Sprintf("%d", itemID)),
				Selected: true,
			})
		}
	}

	if !isUserMenu {
		resources := app.resources
		for _, resourceData := range resources {
			if resourceData.parentBoard != board {
				continue
			}

			if userData.Authorize(resourceData.canView) {
				resourceURL := resourceData.getURL("")
				var selected bool
				if urlPath == resourceURL {
					selected = true
				}

				if selected {
					isExpanded = true
				}

				icon := resourceData.icon

				subitems, expandedSubmenu := resourceData.resourceBoard.getMenuItems(userData, urlPath, false, "")
				if expandedSubmenu {
					isExpanded = true
				}
				if selected {
					expandedSubmenu = true
				}

				ret = append(ret, menuItem{
					Icon:     icon,
					Name:     resourceData.pluralName(userData.Locale()),
					URL:      resourceURL,
					Selected: selected,
					Subitems: subitems,
					Expanded: expandedSubmenu,
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

		icon := v.icon
		if icon == "" {
			icon = iconForm
		}

		menuItem := menuItem{
			Icon:        icon,
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

	if !dontSortByName {
		sortSection(ret, userData.Locale())
	}

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
