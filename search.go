package prago

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/hypertornado/prago/pragelastic"
)

type searchItem struct {
	ID       string
	Icon     string
	Prename  string
	Postname string
	Name     string
	URL      string
}

const searchPageSize int64 = 10

type PaginationItem struct {
	Title    int
	Selected bool
	URL      string
}

var searchClient *pragelastic.Client

func (app *App) createNewElasticSearchClient() error {
	ret, err := pragelastic.New(app.codeName)
	if err != nil {
		return err
	}
	searchClient = ret
	return nil

}

func (app *App) ElasticSearchClient() *pragelastic.Client {
	if searchClient == nil {
		app.createNewElasticSearchClient()
	}
	return searchClient
}

func (app *App) initSearch() {
	db := sysadminBoard.Dashboard(unlocalized("Elasticsearch"))
	db.Task(unlocalized("Reload elasticsearch client")).Handler(func(ta *TaskActivity) error {
		return app.createNewElasticSearchClient()
	})

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

			var pagination []PaginationItem
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
				pagination = append(pagination, PaginationItem{
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

	sysadminBoard.FormAction("delete-elastic-indice", func(f *Form, r *Request) {
		stats, err := app.ElasticSearchClient().GetStats()
		if err != nil {
			panic(err)
		}

		var indiceNames []string
		for k := range stats.Indices {
			indiceNames = append(indiceNames, k)
		}
		sort.Strings(indiceNames)

		var doubled [][2]string = [][2]string{{"", ""}}

		for _, v := range indiceNames {
			doubled = append(doubled, [2]string{v, v})
		}

		f.AddSelect("indice", "Elastic indices", doubled)

		f.AddSubmit("Delete indice")
	}, func(vc ValidationContext) {
		id := vc.GetValue("indice")
		if id == "" {
			vc.AddItemError("indice", "Select indice to delete")
		}
		if !vc.Valid() {
			return
		}

		err := app.ElasticSearchClient().DeleteIndex(id)
		if err != nil {
			vc.AddError(fmt.Sprintf("Index '%s' nelze smazat", id))
		} else {
			vc.AddError(fmt.Sprintf("Index '%s' úspěšně smazán", id))
		}
	}).Name(unlocalized("Smazat elasticsearch index")).Permission(sysadminPermission)

}

func (app *App) searchItems(q string, page int64, request *Request) (ret []*searchItem, hits int64, err error) {
	ret, err = app.getCustomSearch(q, request)
	must(err)

	ret = append(ret, app.searchWithoutElastic(q, request)...)
	hits = int64(len(ret))
	return
}

func (app *App) getCustomSearch(q string, request *Request) (ret []*searchItem, err error) {
	for _, v := range app.customSearchFunctions {
		for _, result := range v(q, request) {
			ret = append(ret, &searchItem{
				Prename: result.Prename,
				Name:    result.Name,
				URL:     result.URL,
			})
		}
	}
	return
}

func (app *App) suggestItems(q string, request *Request) (ret []*searchItem, err error) {
	q = strings.Trim(q, " ")
	if q == "" {
		return ret, nil
	}

	customRes, err := app.getCustomSearch(q, request)
	must(err)
	ret = append(customRes, app.searchWithoutElastic(q, request)...)

	if len(ret) > 5 {
		ret = ret[0:5]
	}

	return
}

func (app *App) searchWithoutElastic(q string, request *Request) (ret []*searchItem) {
	q = normalizeCzechString(q)
	menu := app.getMenu(request, nil)
	for _, item := range menu.Items {
		ret = append(ret, item.SearchWithoutElastic(q, "")...)
	}
	return ret
}

func (item menuItem) SearchWithoutElastic(q string, prename string) (ret []*searchItem) {
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
		ret = append(ret, subitem.SearchWithoutElastic(q, item.Name)...)
	}

	return
}

func (app *App) AddCustomSearchFunction(fn func(q string, userData UserData) []*CustomSearchResult) {
	app.customSearchFunctions = append(app.customSearchFunctions, fn)
}

type CustomSearchResult struct {
	URL     string
	Prename string
	Name    string
}
