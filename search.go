package prago

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/olivere/elastic/v7"
	"golang.org/x/net/context"
)

const searchPageSize int = 10

func (app *App) initSearch() {
	go app.initSearchInner()
}

type searchItem struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	URL         string   `json:"url"`
	Roles       []string `json:"roles"`
}

type searchPage struct {
	Title    int
	Selected bool
	URL      string
}

type adminSearch struct {
	client    *elastic.Client
	app       *App
	indexName string
}

func (si searchItem) CroppedDescription() string {
	return crop(si.Description, 100)
}

func newAdminSearch(app *App) (*adminSearch, error) {
	client, err := elastic.NewClient()
	if err != nil {
		return nil, err
	}
	return &adminSearch{
		client:    client,
		app:       app,
		indexName: "prago_admin",
	}, nil
}

func (e *adminSearch) createSearchIndex() error {
	e.client.DeleteIndex(e.indexName).Do(context.Background())
	e.flush()

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
			"analyzer": "cesky_suggest",
			"preserve_separators": true
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
}

func (e *adminSearch) addItem(item *searchItem, weight int) error {
	var suggest = parseSuggestions(item.Name)
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
	return err
}

func (e *adminSearch) DeleteIndex() error {
	_, err := e.client.DeleteIndex(e.indexName).Do(context.Background())
	return err
}

func (e *adminSearch) Search(q string, role string, page int) ([]*searchItem, int64, error) {
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

	return ret, searchResult.TotalHits(), nil
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

func (e *adminSearch) flush() error {
	_, err := e.client.Flush().Do(context.Background())
	return err
}

func (e *adminSearch) searchImport() error {
	var err error

	err = e.createSearchIndex()
	if err != nil {
		return fmt.Errorf("while creating index: %s", err)
	}

	for _, v := range e.app.resources {
		err = e.importResource(v)
		if err != nil {
			return fmt.Errorf("while importing resource %s: %s", v.id, err)
		}
	}
	e.flush()

	return nil
}

func (e *adminSearch) importResource(resource *resource) error {
	roles := resource.app.getResourceViewRoles(*resource)
	var resourceSearchItem = searchItem{
		ID:    "resource_" + resource.id,
		Name:  resource.name("cs"),
		URL:   resource.getURL(""),
		Roles: roles,
	}
	e.addItem(&resourceSearchItem, 200)

	//var item interface{}
	//resource.newItem(&item)
	c, _ := resource.query().count()
	if c > 10000 {
		return nil
	}

	//var items interface{}
	//resource.newArrayOfItems(&items)
	items, err := resource.query().list()
	if err == nil {
		itemsVal := reflect.ValueOf(items)
		for i := 0; i < itemsVal.Len(); i++ {
			item2 := itemsVal.Index(i).Interface()
			e.saveItemWithRoles(resource, item2, roles)
		}
	}

	return nil
}

func (e *adminSearch) saveItem(resource *resource, item interface{}) error {
	roles := resource.app.getResourceViewRoles(*resource)
	return e.saveItemWithRoles(resource, item, roles)
}

func (e *adminSearch) saveItemWithRoles(resource *resource, item interface{}, roles []string) error {
	//TODO: ugly hack
	relData := resource.itemToRelationData(item, &user{}, nil)
	if relData == nil {
		return errors.New("wrong item to relation data conversion")
	}
	searchItem := relationDataToSearchItem(resource, *relData)
	searchItem.Roles = roles
	return e.addItem(&searchItem, 100)
}

func (e *adminSearch) deleteItem(resource *resource, id int64) error {
	_, err := e.client.Delete().Index(e.indexName).Id(searchID(resource, id)).Do(context.Background())
	return err
}

func searchID(resource *resource, id int64) string {
	return fmt.Sprintf("%s-%d", resource.id, id)
}

func relationDataToSearchItem(resource *resource, data viewRelationData) searchItem {
	return searchItem{
		ID:          searchID(resource, data.ID),
		Category:    resource.name("cs"),
		Name:        data.Name,
		Description: data.Description,
		Image:       data.Image,
		URL:         data.URL,
	}

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

	app.Action("_search").Permission(loggedPermission).Name(unlocalized("Vyhledávání")).Template("admin_search").IsWide().hiddenMenu().DataSource(
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

			result, hits, err := adminSearch.Search(q, request.user.Role, page-1)
			must(err)

			var pages = int(hits) / searchPageSize
			if hits > 0 {
				pages++
			}

			var searchPages []searchPage
			for i := 1; i <= pages; i++ {
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

func parseSuggestions(in string) []string {
	parts := strings.Split(in, " ")
	ret := []string{}
	for i := 0; i < len(parts); i++ {
		subparts := parts[i:]
		add := strings.Join(subparts, " ")
		if add != "" {
			ret = append(ret, strings.Join(subparts, " "))
		}
	}
	return ret
}
