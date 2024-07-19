package prago

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

type menu struct {
	Language    string
	SearchQuery string
	Items       []*menuItem

	Username           string
	Email              string
	Role               string
	RoleWarning        bool
	LanguageDecription string
	Version            string
}

type menuItem struct {
	Icon         string
	Image        string
	Name         string
	URL          string
	Subitems     []*menuItem
	Selected     bool
	Expanded     bool
	SortPriority int64
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

func (app *App) getMenu(request *Request, item any) (ret *menu) {

	menuContext := getMenuRequestContextFromRequest(request, item)

	ret = &menu{
		Items: app.MainBoard.getMenuItems(menuContext),
	}
	ret.Language = request.Locale()

	user := request.getUser()
	ret.Username = fmt.Sprintf("Přihlášený uživatel %s %s", user.Username, user.Name)
	ret.Email = user.Email
	if request.role() != "" {
		ret.Role = fmt.Sprintf("Role „%s“", request.role())
	} else {
		ret.RoleWarning = true
		ret.Role = "Nebyla vám zatím administrátorem webu přidělena žádná role"
	}
	ret.LanguageDecription = fmt.Sprintf("Jazyk %s", localeNames[user.Locale])
	ret.Version = "Verze " + app.version
	return ret
}

func (menu menu) GetIcon() string {
	return getIconFromMenuSubsections(menu.Items)
}

func getIconFromMenuSubsections(items []*menuItem) string {
	for _, v := range items {
		if v.Selected {
			return v.Icon
		}
		icon := getIconFromMenuSubsections(v.Subitems)
		if icon != "" {
			return icon
		}
	}
	return ""
}

func (menu menu) GetTitle() string {
	for _, item := range menu.Items {
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

func getTitleFromMenuSubsections(item *menuItem) []string {
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

func (board *Board) getMenuItems(requestContext *menuRequestContext) []*menuItem {
	urlPath := requestContext.URL
	csrfToken := requestContext.CSRFToken

	app := board.app
	var ret []*menuItem

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

			ret = append(ret, &menuItem{
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
		if icon == "" {
			icon = iconForm
		}

		menuItem := &menuItem{
			Icon:         icon,
			Name:         v.name(requestContext.UserData.Locale()),
			URL:          fullURL,
			Selected:     selected,
			SortPriority: sortPriority,
		}

		if v.isPartOfBoard != nil && v.isPartOfBoard != app.MainBoard {
			menuItem.Subitems = v.isPartOfBoard.getMenuItems(requestContext)
		}

		ret = append(ret, menuItem)

	}
	sortAndExpandMenuItems(ret)
	return ret
}

func (resource *Resource) getResourceMenu(requestContext *menuRequestContext) (ret []*menuItem) {
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
		menuItem := &menuItem{
			Icon:         v.icon,
			Name:         v.name(requestContext.UserData.Locale()),
			URL:          resource.getURL(v.url),
			SortPriority: v.priority - int64(k),
		}
		if urlPath == menuItem.URL {
			menuItem.Selected = true
		}

		if v.url == "list" && resource.isItPointerToResourceItem(requestContext.Item) {
			menuItem.Subitems = append(menuItem.Subitems, resource.getResourceItemMenu(requestContext))
			//menuItem.Expanded = true
		}
		ret = append(ret, menuItem)
	}

	sortAndExpandMenuItems(ret)
	return
}

func (resource *Resource) getResourceItemMenu(requestContext *menuRequestContext) *menuItem {
	var items []*menuItem

	for k, v := range resource.itemActions {
		if v.method != "GET" {
			continue
		}
		if !requestContext.UserData.Authorize(v.permission) {
			continue
		}
		name := v.name(requestContext.UserData.Locale())
		var thumbnail string
		if v.url == "" {
			previewer := resource.previewer(requestContext.UserData, requestContext.Item)
			thumbnail = previewer.ThumbnailURL()
			name = previewer.Name()
		}

		priority := v.priority - int64(k)

		item := &menuItem{
			Icon:         v.icon,
			Image:        thumbnail,
			Name:         name,
			URL:          resource.getItemURL(requestContext.Item, v.url, requestContext.UserData),
			Expanded:     true,
			SortPriority: priority,
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

func sortAndExpandMenuItems(items []*menuItem) {
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

func sortSection(items []*menuItem) {
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

func (item *menuItem) IsSelectedOrExpanded() bool {
	return item.Selected || item.Expanded
}
