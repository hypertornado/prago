package prago

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hypertornado/prago/pragelastic"
	//"github.com/olivere/elastic"
)

type searchItem struct {
	ID          string
	Category    string `elastic-datatype:"keyword"`
	Name        string `elastic-datatype:"text" elastic-analyzer:"czech"`
	Description string `elastic-datatype:"text" elastic-analyzer:"czech"`
	Image       string `elastic-datatype:"keyword"`
	URL         string `elastic-datatype:"keyword"`
	Roles       []string
}

const searchPageSize int64 = 10

type searchPage struct {
	Title    int
	Selected bool
	URL      string
}

func (app *App) initSearch() {
	go app.initSearchInner()
}

func (app *App) initSearchInner() {
	var err error

	adminSearch, err := newAdminSearch(app)
	if err != nil {
		if app.developmentMode {
			app.Log().Println("admin search not initialized: " + err.Error())
		}
		return
	}
	if app.developmentMode {
		app.Log().Println("admin search initialized")
	}
	app.search = adminSearch

	go func() {
		err := adminSearch.searchImport()
		if err != nil {
			app.Log().Println(fmt.Errorf("%s", err))
		}
	}()

	app.Action("_search").Permission(loggedPermission).Name(unlocalized("Vyhledávání")).Template("admin_search").hiddenInMainMenu().DataSource(
		func(request *Request) interface{} {
			q := request.Params().Get("q")
			pageStr := request.Params().Get("page")

			var page = 1
			if pageStr != "" {
				var err error
				page, err = strconv.Atoi(pageStr)
				if err != nil {
					panic("no search parameter")
				}
			}

			result, hits, err := adminSearch.Search(q, request.user.Role, int64(page-1))
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

			title := fmt.Sprintf("Vyhledávání – \"%s\" – %d výsledků", q, hits)

			var ret = map[string]interface{}{}

			request.SetData("search_q", q)

			ret["search_q"] = q
			ret["admin_title"] = title
			ret["search_results"] = result
			ret["search_pages"] = searchPages

			return ret
		},
	)

	app.API("search-suggest").Permission(loggedPermission).Handler(
		func(request *Request) {
			results, err := adminSearch.Suggest(request.Params().Get("q"))
			if err != nil {
				app.Log().Println(err)
			}

			if len(results) == 0 {
				request.RenderJSONWithCode(nil, 204)
				return
			}

			request.SetData("items", results)
			request.RenderView("admin_search_suggest")
		},
	)
}

type adminSearch struct {
	client *pragelastic.Client
	app    *App
	index  *pragelastic.Index[searchItem]
}

func newAdminSearch(app *App) (*adminSearch, error) {
	client, err := pragelastic.New(app.codeName)
	if err != nil {
		return nil, err
	}
	index := pragelastic.NewIndex[searchItem](client)

	return &adminSearch{
		client: client,
		app:    app,
		index:  index,
	}, nil
}

func (resource *Resource[T]) importSearchData(bulkUpdater *pragelastic.BulkUpdater[searchItem]) error {
	roles := resource.getResourceViewRoles()
	var resourceSearchItem = searchItem{
		ID:    "resource_" + resource.id,
		Name:  resource.getPluralNameFunction()("cs"),
		URL:   resource.getURL(""),
		Roles: roles,
	}

	resource.app.search.addItem(bulkUpdater, &resourceSearchItem, 200)

	c, _ := resource.Query().Count()
	if c > 10000 {
		return nil
	}

	items := resource.Query().List()
	for _, item := range items {
		resource.saveSearchItemWithRoles(bulkUpdater, item, roles)
	}
	return nil
}

func (e *adminSearch) flush() error {
	return e.index.Flush()
}

func (resource *Resource[T]) saveSearchItem(item *T) error {
	roles := resource.getResourceViewRoles()
	return resource.saveSearchItemWithRoles(nil, item, roles)
}

func (resource *Resource[T]) saveSearchItemWithRoles(bulkUpdater *pragelastic.BulkUpdater[searchItem], item *T, roles []string) error {
	//TODO: ugly hack
	preview := resource.getPreview(item, &user{}, nil)
	if preview == nil {
		return errors.New("wrong item to relation data conversion")
	}
	searchItem := relationDataToSearchItem(resource, *preview)
	searchItem.Roles = roles
	return resource.app.search.addItem(bulkUpdater, &searchItem, 100)
}

func relationDataToSearchItem(resource resourceIface, data preview) searchItem {
	return searchItem{
		ID:          searchID(resource, data.ID),
		Category:    resource.getPluralNameFunction()("cs"),
		Name:        data.Name,
		Description: data.Description,
		Image:       data.Image,
		URL:         data.URL,
	}
}

func searchID(resource resourceIface, id int64) string {
	return fmt.Sprintf("%s-%d", resource.getID(), id)
}

func (e *adminSearch) deleteItem(resource resourceIface, id int64) error {
	return e.index.DeleteItem(searchID(resource, id))
}

func (e *adminSearch) searchImport() error {
	var err error

	bulkUpdater, err := e.index.UpdateBulk()
	if err != nil {
		return fmt.Errorf("can't create bulk update: %s", err)
	}

	err = e.createSearchIndex()
	if err != nil {
		return fmt.Errorf("while creating index: %s", err)
	}

	for _, v := range e.app.resources {
		total, _ := e.index.Count()
		fmt.Printf("importing resource: %s, total: %d\n", v.getID(), total)
		err = v.importSearchData(bulkUpdater)
		if err != nil {
			return fmt.Errorf("while importing resource %s: %s", v.getID(), err)
		}
	}

	err = bulkUpdater.Close()
	if err != nil {
		return err
	}
	err = e.index.Flush()
	if err != nil {
		return err
	}
	err = e.index.Refresh()
	if err != nil {
		return err
	}
	fmt.Println("INDEX Created")

	return nil
}

func (e *adminSearch) createSearchIndex() error {
	e.index.Delete()
	err := e.index.Create()
	if err != nil {
		return err
	}

	//pragelastic.New("xxx")

	//e.client

	/*

			e.client.DeleteIndex(e.indexName).Do(context.Background())
			e.flush()

			//e.client.CreateIndex(e.indexName).

			_, err := e.client.CreateIndex(e.indexName).BodyString(`
		    {
		      "settings": {
		          "analysis": {
		            "filter": {
		              "czech_stop": {
		                "type":       "stop",
		                "stopwords":  "_czech_"
		              },
		              "czech_keywords": {
		                "type":       "keyword_marker",
		                "keywords":   ["a"]
		              },
		              "czech_stemmer": {
		                "type":       "stemmer",
		                "language":   "czech"
		              }
		            },
		            "analyzer": {
		              "cesky": {
		                "tokenizer":  "standard",
		                "filter": [
		                  "lowercase",
		                  "asciifolding",
		                  "czech_stop",
		                  "czech_keywords",
		                  "czech_stemmer"
		                ]
		              },
		              "cesky_suggest": {
		                "tokenizer":  "standard",
		                "filter": [
		                  "lowercase",
		                  "asciifolding"
		                ]
		              }
		            }
		          }
		        },
		      "mappings": {
				"properties": {
				"suggest": {
					"type": "completion",
					"analyzer": "cesky_suggest"
				},
				"name": {"type": "text", "analyzer": "cesky"},
				"description": {"type": "text", "analyzer": "cesky"},
				"image": {"type": "text"},
							"url": {"type": "text"},
							"roles": {"type": "text"}
				}
		      }
		    }
				`).Do(context.Background())
			if err != nil {
				return fmt.Errorf("while creating index %s", err)
			}
			return nil
	*/
	return nil
}

func (e *adminSearch) addItem(bulkUpdater *pragelastic.BulkUpdater[searchItem], item *searchItem, weight int) error {
	if bulkUpdater != nil {
		bulkUpdater.AddItem(item)
		return nil
	}
	return e.index.UpdateSingle(item)
	/*var suggest = parseSuggestions(item.Name)
	_, err := e.client.Index().Index(e.indexName).BodyJson(map[string]interface{}{
		"suggest": map[string]interface{}{
			"input":  suggest,
			"weight": weight,
		},
		"category":    item.Category,
		"name":        item.Name,
		"description": item.Description,
		"image":       item.Image,
		"url":         item.URL,
		"roles":       item.Roles,
	}).Id(item.ID).Do(context.Background())
	return err*/
	//return nil
}

func (e *adminSearch) DeleteIndex() error {
	return e.index.Delete()
}

func (e *adminSearch) Search(q string, role string, page int64) ([]*searchItem, int64, error) {
	items, totalHits, err := e.index.Query().
		Offset(page*searchPageSize).
		Limit(searchPageSize).
		Filter("Name", q).
		List()
	if err != nil {
		return nil, -1, err
	}
	//fmt.Println(items)
	return items, totalHits, nil
	/*

		var ret []*searchItem

		mq := elastic.NewMultiMatchQuery(q)
		mq.FieldWithBoost("name", 3)
		mq.FieldWithBoost("description", 2)

		bq := elastic.NewBoolQuery()
		bq.Must(
			elastic.NewTermsQuery("roles", role),
		)
		bq.Must(mq)

		searchResult, err := e.client.Search().
			Index(e.indexName).
			Query(bq).
			From(page * searchPageSize).
			Size(searchPageSize).
			Do(context.Background())
		if err != nil {
			return nil, 0, err
		}

		var item searchItem
		for _, item := range searchResult.Each(reflect.TypeOf(item)) {
			if t, ok := item.(searchItem); ok {
				ret = append(ret, &t)
			}
		}

		return ret, searchResult.TotalHits(), nil*/
}

//TODO: add roles to suggest
func (e *adminSearch) Suggest(q string) ([]*searchItem, error) {
	//disabled
	return []*searchItem{}, nil
	/*

		suggesterName := "completion_suggester"
		cs := elastic.NewCompletionSuggester(suggesterName)
		cs = cs.Field("suggest")
		cs = cs.Prefix(q)
		cs = cs.SkipDuplicates(true)

		searchResult, err := e.client.Search().
			Index(e.indexName).
			Suggester(cs).
			Pretty(true).
			Do(context.Background())
		if err != nil {
			return nil, err
		}

		suggestions := searchResult.Suggest[suggesterName]

		var ret []*searchItem

		multi := e.client.MultiGet()
		for _, v := range suggestions {
			for _, v2 := range v.Options {
				multiitem := elastic.NewMultiGetItem().Id(v2.Id).Index(e.indexName)
				multi = multi.Add(multiitem)
			}
		}

		res, err := multi.Do(context.Background())
		if err != nil {
			return nil, err
		}
		for _, v := range res.Docs {
			if v.Source != nil {
				var item searchItem
				err = json.Unmarshal(v.Source, &item)
				if err == nil {
					ret = append(ret, &item)
				}
			}
		}

		return ret, nil*/
}
