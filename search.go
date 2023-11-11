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
	ID   string
	Name string
	URL  string
}

const searchPageSize int64 = 10

type searchPage struct {
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

	app.Action("_search").Permission(loggedPermission).Name(unlocalized("Vyhledávání")).Template("admin_search").hiddenInMenu().DataSource(
		func(request *Request) interface{} {
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

			var searchPages []searchPage
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
				searchPages = append(searchPages, searchPage{
					Title:    i,
					Selected: selected,
					URL:      "_search?" + values.Encode(),
				})
			}

			var ret = map[string]interface{}{}

			ret["box_header"] = BoxHeader{
				Name:      fmt.Sprintf("Vyhledávání – „%s", q),
				TextAfter: fmt.Sprintf("%s výsledků", humanizeNumber(hits)),
			}
			ret["search_q"] = q
			ret["admin_title"] = fmt.Sprintf("„%s", q)
			ret["search_results"] = result
			ret["search_pages"] = searchPages

			return ret
		},
	)

	sysadminBoard.FormAction("delete-elastic-indice").Name(unlocalized("Smazat elasticsearch index")).Permission(sysadminPermission).Form(func(f *Form, r *Request) {
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
	}).Validation(func(vc ValidationContext) {
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
	})

}

func (app *App) searchItems(q string, page int64, request *Request) (ret []*searchItem, hits int64, err error) {
	ret = app.searchWithoutElastic(q, request)
	hits = int64(len(ret))
	return
}

func (app *App) suggestItems(q string, request *Request) (ret []*searchItem, err error) {
	q = strings.Trim(q, " ")
	if q == "" {
		return ret, nil
	}

	ret = app.searchWithoutElastic(q, request)

	if len(ret) > 5 {
		ret = ret[0:5]
	}

	return
}

func (app *App) searchWithoutElastic(q string, request *Request) (ret []*searchItem) {
	q = normalizeCzechString(q)
	menu := app.getMenu(request, "", "")
	for _, section := range menu.Sections {
		for _, item := range section.Items {
			ret = append(ret, item.SearchWithoutElastic(q)...)
		}
	}
	return ret
}

func (item menuItem) SearchWithoutElastic(q string) (ret []*searchItem) {
	if strings.HasPrefix(item.URL, "/admin/logout") {
		return
	}
	name := normalizeCzechString(item.Name)
	if strings.Contains(name, q) {
		ret = append(ret, &searchItem{
			Name: item.Name,
			URL:  item.URL,
		})
	}

	for _, subitem := range item.Subitems {
		ret = append(ret, subitem.SearchWithoutElastic(q)...)
	}

	return
}
