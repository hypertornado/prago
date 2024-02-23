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

const searchPageSize int64 = 10

type paginationItem struct {
	Title    int
	Selected bool
	URL      string
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
			request.WriteHTML(200, "admin_search_suggest", results)
		},
	)

	app.Action("_search").Permission(loggedPermission).Name(unlocalized("Vyhledávání")).Board(nil).ui(
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

			result, hits, err := app.searchItems(q, int64(page-1), request)
			must(err)

			var pages = hits / searchPageSize
			if hits > 0 {
				pages++
			}

			var pagination []paginationItem
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
				pagination = append(pagination, paginationItem{
					Title:    i,
					Selected: selected,
					URL:      "_search?" + values.Encode(),
				})
			}

			view := &view{
				Header: &boxHeader{
					Name:      fmt.Sprintf("Vyhledávání – „%s“", q),
					TextAfter: fmt.Sprintf("%s výsledků", humanizeNumber(hits)),
				},

				SearchResults: result,
				Pagination:    pagination,
			}

			pd.Name = view.Header.Name

			pd.Views = append(pd.Views, view)

			pd.SearchQuery = q
		},
	)
}

func (app *App) searchItems(q string, page int64, request *Request) (ret []*searchItem, hits int64, err error) {
	/*ret, err = app.getCustomSearch(q, request)
	must(err)*/

	ret = append(ret, app.searchWithoutElastic(q, request)...)
	hits = int64(len(ret))
	return
}

/*func (app *App) getCustomSearch(q string, request *Request) (ret []*searchItem, err error) {
	for _, v := range app.customSearchFunctions {
		for _, result := range v(q, request) {
			ret = append(ret, &searchItem{
				Prename:  result.Prename,
				Name:     result.Name,
				URL:      result.URL,
				Priority: result.Priority,
			})
		}
	}
	return
}*/

func (app *App) suggestItems(q string, request *Request) (ret []*searchItem, err error) {
	q = strings.Trim(q, " ")
	if q == "" {
		return ret, nil
	}

	//customRes, err := app.getCustomSearch(q, request)
	//must(err)
	ret = append(ret, app.searchWithoutElastic(q, request)...)

	if len(ret) > 5 {
		ret = ret[0:5]
	}

	return
}

func (app *App) searchWithoutElastic(q string, request *Request) (ret []*searchItem) {
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

	return ret
}

func (item menuItem) searchMenuItem(q string, prename string) (ret []*searchItem) {
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
