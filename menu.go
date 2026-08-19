package prago

import (
	"fmt"
	"html/template"
	"sort"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type MenuItem struct {
	Icon         string
	Image        string
	Name         string
	URL          string
	Subitems     []*MenuItem
	Selected     bool
	Expanded     bool
	SortPriority int64
	NoSearch     bool
	Style        string
}

type menuRequestContext struct {
	URL       string
	UserData  UserData
	Item      any
	CSRFToken string
}

func getMenuRequestContextFromRequest(request *Request, item any) *menuRequestContext {
	ret := &menuRequestContext{
		URL:       request.Request().URL.Path,
		UserData:  request,
		Item:      item,
		CSRFToken: request.csrfToken(),
	}
	return ret
}

func (app *App) getFooterItems(request *Request) (ret []template.HTML) {
	user := request.getUser()
	ret = append(ret, template.HTML(fmt.Sprintf("%s %s", user.Username, user.Name)))
	ret = append(ret, template.HTML(user.Email))

	var items []string
	items = append(items, app.codeName)
	items = append(items, fmt.Sprintf("%s", app.version))

	if request.Role() != "" {
		roleID := request.Role()
		roleName := roleID
		if request.app.accessManager.roleNames[roleID] != nil {
			roleName = request.app.accessManager.roleNames[roleID](request.Locale())
		}
		items = append(items, roleName)
	} else {
		items = append(items, "Nebyla vám zatím administrátorem webu přidělena žádná role")
	}

	items = append(items, localeNames[user.Locale])
	ret = append(ret, template.HTML(strings.Join(items, " · ")))
	return ret

}

func (app *App) getMenuItems(request *Request, item any) (ret []*MenuItem) {
	menuContext := getMenuRequestContextFromRequest(request, item)
	return app.MainBoard.getMenuItems(menuContext)
}

func (pd *PageData) GetIconAndStyle() (string, string) {
	return getIconFromMenuSubsections(pd.MenuItems)
}

func getIconFromMenuSubsections(items []*MenuItem) (string, string) {
	for _, v := range items {
		if v.Selected {
			return v.Icon, v.Style
		}
		icon, style := getIconFromMenuSubsections(v.Subitems)
		if icon != "" {
			return icon, style
		}
	}
	return "", ""
}

func (pd *PageData) GetTitle() string {
	for _, item := range pd.MenuItems {
		if item.Selected {
			return item.Name
		}
		ret := getTitleFromMenuSubsections(item)
		if len(ret) > 0 {
			return strings.Join(ret, " · ")
		}
	}
	return ""
}

func getTitleFromMenuSubsections(item *MenuItem) []string {
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
	app.NewAPI("resource-counts").Permission(loggedPermission).HandlerJSON(func(request *Request) any {
		return getResourceCountsMap(request)
	})
}

func getResourceCountsMap(request *Request) map[string]string {
	app := request.app
	ret := make(map[string]string)

	for _, v := range app.resources {
		if request.Authorize(v.canView) {
			url := v.getURL("")
			count := v.getCachedCount()
			ret[url] = humanizeNumber(count)
		}

	}
	return ret
}

const sortPriorityBoard = 10
const sortPriorityMainBoard = 20

func (board *Board) getMenuItems(requestContext *menuRequestContext) []*MenuItem {
	urlPath := requestContext.URL
	csrfToken := requestContext.CSRFToken

	app := board.app
	var ret []*MenuItem

	if board.parentResource != nil {
		ret = board.parentResource.getResourceMenu(requestContext)
	}

	resources := app.resources
	for _, resource := range resources {
		if resource.parentBoard != board {
			continue
		}

		if requestContext.UserData.Authorize(resource.canView) {
			resourceURL := resource.getURL("")
			var selected bool
			if urlPath == resourceURL {
				selected = true
			}

			icon := resource.icon

			subitems := resource.resourceBoard.getMenuItems(requestContext)

			ret = append(ret, &MenuItem{
				Icon:         icon,
				Name:         resource.pluralName(requestContext.UserData.Locale()),
				URL:          resourceURL,
				Selected:     selected,
				Subitems:     subitems,
				SortPriority: 10,
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
		if !requestContext.UserData.Authorize(v.permission) {
			continue
		}

		var selected bool
		fullURL := app.getAdminURL(v.url)
		if urlPath == fullURL {
			selected = true
		}

		var sortPriority int64
		if fullURL == "/admin/logout" {
			sortPriority = -1
			fullURL += "?_csrfToken=" + csrfToken
		}

		if v.isPartOfBoard != nil {
			if v.isPartOfBoard.isEmpty(requestContext) {
				continue
			}
			sortPriority = sortPriorityBoard
			if v.isPartOfBoard.isMainBoard() {
				sortPriority = sortPriorityMainBoard
			}
		}

		icon := v.icon

		if fullURL == "/admin/_options" {
			sortPriority = -1
		}

		menuItem := &MenuItem{
			Icon:         icon,
			Name:         v.name(requestContext.UserData.Locale()),
			URL:          fullURL,
			Selected:     selected,
			SortPriority: sortPriority,
			Style:        v.style,
		}

		if v.isPartOfBoard != nil && v.isPartOfBoard != app.MainBoard {
			menuItem.Subitems = v.isPartOfBoard.getMenuItems(requestContext)
		}

		ret = append(ret, menuItem)

	}
	sortAndExpandMenuItems(ret)
	return ret
}

func (resource *Resource) getResourceMenu(requestContext *menuRequestContext) (ret []*MenuItem) {
	urlPath := requestContext.URL
	for k, v := range resource.actions {
		if v.method != "GET" {
			continue
		}
		if !requestContext.UserData.Authorize(v.permission) {
			continue
		}
		if v.url == "" {
			continue
		}
		menuItem := &MenuItem{
			Icon:         v.icon,
			Name:         v.name(requestContext.UserData.Locale()),
			URL:          resource.getURL(v.url),
			SortPriority: v.priority - int64(k),
			NoSearch:     true,
			Style:        v.style,
		}
		if urlPath == menuItem.URL {
			menuItem.Selected = true
		}

		if v.url == "list" && resource.isItPointerToResourceItem(requestContext.Item) {
			menuItem.Subitems = append(menuItem.Subitems, resource.getResourceItemMenu(requestContext))
		}
		ret = append(ret, menuItem)
	}

	sortAndExpandMenuItems(ret)
	return
}

func (resource *Resource) getResourceItemMenu(requestContext *menuRequestContext) *MenuItem {
	var items []*MenuItem

	for k, v := range resource.itemActions {
		if v.method != "GET" {
			continue
		}
		if !requestContext.UserData.Authorize(v.permission) {
			continue
		}
		if v.isFormMultipleAction {
			continue
		}

		icon := v.icon
		style := v.style
		name := v.name(requestContext.UserData.Locale())
		var thumbnail string
		if v.url == "" {
			previewer := resource.previewer(requestContext.UserData, requestContext.Item)
			thumbnail = previewer.ThumbnailURL()
			name = previewer.Name()
			if icon == "" {
				icon = previewer.Icon()
			}
			if style == "" {
				style = previewer.Style()
			}
		}

		priority := v.priority - int64(k)

		item := &MenuItem{
			Icon:         icon,
			Image:        thumbnail,
			Name:         name,
			URL:          resource.getItemURL(requestContext.Item, v.url, requestContext.UserData),
			Expanded:     true,
			SortPriority: priority,
			NoSearch:     true,
			Style:        style,
		}

		if requestContext.URL == item.URL {
			item.Selected = true
		}
		items = append(items, item)
	}

	if len(items) == 0 {
		return nil
	}
	sortAndExpandMenuItems(items)
	ret := items[0]
	ret.Subitems = items[1:]
	return ret

}

func sortAndExpandMenuItems(items []*MenuItem) {
	sortSection(items)
	for _, item := range items {
		var expanded bool
		for _, subitem := range item.Subitems {
			if subitem.Expanded || subitem.Selected {
				expanded = true
			}
		}
		if !item.Expanded {
			item.Expanded = expanded
		}
	}
}

func sortSection(items []*MenuItem) {
	collator := collate.New(language.Czech)

	sort.SliceStable(items, func(i, j int) bool {
		a := items[i]
		b := items[j]

		if a.SortPriority > b.SortPriority {
			return true
		}
		if a.SortPriority < b.SortPriority {
			return false
		}

		if collator.CompareString(a.Name, b.Name) <= 0 {
			return true
		} else {
			return false
		}
	})
}

func (item *MenuItem) IsSelectedOrExpanded() bool {
	return item.Selected || item.Expanded
}
