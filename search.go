package prago

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type searchItem struct {
	ID       string
	Icon     string
	Prename  string
	Postname string
	Name     string
	URL      string
	Priority int64
}

const searchPageSize int64 = 20

type paginationItem struct {
	Title    int
	Selected bool
	URL      string
}

func searchItemsToViewField(items []*searchItem) *viewField {
	ret := &viewField{}

	ret.ViewContent = &viewFieldContent{}

	for _, item := range items {

		preview := &Preview{
			Name:        item.Name,
			Description: item.Prename,
			Icon:        item.Icon,
			URL:         item.URL,
		}

		ret.ViewContent.Previews = append(ret.ViewContent.Previews, preview)
	}

	return ret
}

func (app *App) initSearch() {

	app.API("search-suggest").Permission(loggedPermission).Handler(
		func(request *Request) {
			results, err := app.suggestItems(request.Param("q"), request)
			must(err)
			if len(results) == 0 {
				request.WriteJSON(204, nil)
				return
			}
			request.WriteHTML(200, request.app.adminTemplates, "search_suggest", results)
		},
	)

	ActionPlain(app, "_search", nil).Permission(loggedPermission).Name(unlocalized("Vyhledávání")).Board(nil).ui(
		func(request *Request, pd *pageData) {
			q := request.Param("q")
			pageStr := request.Param("page")

			var page = 1
			if pageStr != "" {
				var err error
				page, err = strconv.Atoi(pageStr)
				if err != nil {
					panic("no search parameter")
				}
			}

			result, hits, err := app.search(q, int64(page-1)*searchPageSize, searchPageSize, request)
			must(err)

			var pages = hits / searchPageSize
			if hits > 0 {
				pages++
			}

			var pagination []*paginationItem
			for i := 1; i <= int(pages); i++ {
				var selected bool
				if page == i {
					selected = true
				}
				values := make(url.Values)
				values.Add("q", q)
				if i > 0 {
					values.Add("page", strconv.Itoa(i))
				}
				pagination = append(pagination, &paginationItem{
					Title:    i,
					Selected: selected,
					URL:      "_search?" + values.Encode(),
				})
			}

			pd.BoxHeader = &boxHeader{
				DescriptionsBefore: []string{"Vyhledávání"},
				Icon:               "glyphicons-basic-28-search.svg",
				Name:               fmt.Sprintf("„%s“", q),
				DescriptionsAfter:  []string{fmt.Sprintf("%s výsledků", humanizeNumber(hits))},
			}

			pd.ViewFields = append(pd.ViewFields, searchItemsToViewField(result))

			pd.Pagination = pagination

			pd.Name = pd.BoxHeader.Name

			pd.SearchQuery = q
		},
	)
}

func (app *App) suggestItems(q string, request *Request) (ret []*searchItem, err error) {
	q = strings.Trim(q, " ")
	if q == "" {
		return ret, nil
	}

	var suggestLimit int64 = 20

	suggestItems, _, err := app.search(q, 0, suggestLimit, request)
	if err != nil {
		return nil, err
	}

	ret = append(ret, suggestItems...)

	return
}

func (app *App) search(q string, offset int64, limit int64, request *Request) (ret []*searchItem, hits int64, err error) {

	if offset < 0 {
		return nil, 0, fmt.Errorf("offset must be non negative")
	}

	if limit <= 0 {
		return nil, 0, fmt.Errorf("limit must be positive")
	}

	q = normalizeCzechString(q)
	menu := app.getMenu(request, nil)
	for _, item := range menu.Items {
		ret = append(ret, item.searchMenuItem(q, "")...)
	}

	for _, fn := range app.customSearchFunctions {
		customSearchResults := fn(q, request)
		for _, result := range customSearchResults {
			ret = append(ret, &searchItem{
				Prename:  result.Prename,
				Name:     result.Name,
				URL:      result.URL,
				Priority: result.Priority,
			})
		}

	}

	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].Priority < ret[j].Priority
	})

	hits = int64(len(ret))

	if offset >= hits {
		return nil, hits, nil
	}

	end := offset + limit
	if end > hits {
		end = hits
	}
	ret = ret[offset:end]

	return ret, hits, nil
}

func (item menuItem) searchMenuItem(q string, prename string) (ret []*searchItem) {
	if item.NoSearch {
		return
	}

	if strings.HasPrefix(item.URL, "/admin/logout") {
		return
	}
	name := normalizeCzechString(item.Name)
	if strings.Contains(name, q) {
		ret = append(ret, &searchItem{
			Icon:    item.Icon,
			Prename: prename,
			Name:    item.Name,
			URL:     item.URL,
		})
	}

	for _, subitem := range item.Subitems {
		ret = append(ret, subitem.searchMenuItem(q, item.Name)...)
	}

	return
}

func (app *App) AddCustomSearchFunction(fn func(q string, userData UserData) []*CustomSearchResult) {
	app.customSearchFunctions = append(app.customSearchFunctions, fn)
}

type CustomSearchResult struct {
	URL      string
	Prename  string
	Name     string
	Priority int64
}

func AddResourceCustomSearchFunction[T any](app *App, fn func(q string, userData UserData) []*T) {
	resource := getResource[T](app)
	resource.customSearchFunctions = append(resource.customSearchFunctions,
		func(q string, userData UserData) (ret []*Preview) {
			items := fn(q, userData)
			for _, item := range items {
				preview := resource.previewer(userData, item).Preview(nil)
				ret = append(ret, preview)
			}
			return ret
		},
	)
}
