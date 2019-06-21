package administration

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/hypertornado/prago"
	"github.com/olivere/elastic"
	"golang.org/x/net/context"
)

const searchPageSize int = 10

const searchType string = "items"

type SearchItem struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	URL         string   `json:"url"`
	Roles       []string `json:"roles"`
}

type SearchPage struct {
	Title    int
	Selected bool
	URL      string
}

type adminSearch struct {
	client    *elastic.Client
	admin     *Administration
	indexName string
}

func NewAdminSearch(admin *Administration) (*adminSearch, error) {
	client, err := elastic.NewClient()
	if err != nil {
		return nil, err
	}
	return &adminSearch{
		client:    client,
		admin:     admin,
		indexName: "prago_admin",
	}, nil
}

func (e *adminSearch) createSearchIndex() error {
	e.client.DeleteIndex(e.indexName).Do(context.Background())
	e.Flush()

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
        "items": {
          "_all": {},
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
    }
		`).Do(context.Background())
	if err != nil {
		return fmt.Errorf("while creating index %s", err)
	}
	return nil
}

func (e *adminSearch) AddItem(item *SearchItem, weight int) error {
	var suggest = parseSuggestions(item.Name)
	_, err := e.client.Index().Index(e.indexName).Type(searchType).BodyJson(map[string]interface{}{
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

func (e *adminSearch) Search(q string, role string, page int) ([]*SearchItem, int64, error) {
	var ret []*SearchItem

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
		Type(searchType).
		Query(bq).
		From(page * searchPageSize).
		Size(searchPageSize).
		Do(context.Background())
	if err != nil {
		return nil, 0, err
	}

	var item SearchItem
	for _, item := range searchResult.Each(reflect.TypeOf(item)) {
		if t, ok := item.(SearchItem); ok {
			ret = append(ret, &t)
		}
	}

	return ret, searchResult.TotalHits(), nil
}

func (e *adminSearch) Suggest(q string) ([]*SearchItem, error) {
	suggesterName := "completion_suggester"
	cs := elastic.NewCompletionSuggester(suggesterName)
	cs = cs.Field("suggest")
	cs = cs.Prefix(q)
	cs = cs.SkipDuplicates(true)

	searchResult, err := e.client.Search().
		Index(e.indexName).
		Type(searchType).
		Suggester(cs).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	suggestions := searchResult.Suggest[suggesterName]

	var ret []*SearchItem

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
			var item SearchItem
			err = json.Unmarshal(*v.Source, &item)
			if err == nil {
				ret = append(ret, &item)
			}
		}
	}

	return ret, nil
}

func (e *adminSearch) Flush() error {
	_, err := e.client.Flush().Do(context.Background())
	return err
}

func (e *adminSearch) searchImport() error {
	fmt.Println("Importing admin search...")
	var err error

	err = e.createSearchIndex()
	if err != nil {
		return fmt.Errorf("while creating index: %s", err)
	}

	for _, v := range e.admin.Resources {
		err = e.importResource(v)
		if err != nil {
			return fmt.Errorf("while importing resource %s: %s", v.TableName, err)
		}
	}
	e.Flush()

	return nil
}

func (e *adminSearch) importResource(resource *Resource) error {

	roles := resource.Admin.getResourceViewRoles(*resource)
	var resourceSearchItem = SearchItem{
		ID:    "resource_" + resource.ID,
		Name:  resource.HumanName("cs"),
		URL:   resource.GetURL(""),
		Roles: roles,
	}
	e.AddItem(&resourceSearchItem, 200)

	var item interface{}
	resource.newItem(&item)
	c, _ := e.admin.Query().Count(item)
	if c > 10000 {
		return nil
	}

	var items interface{}
	resource.newArrayOfItems(&items)
	err := e.admin.Query().Get(items)
	if err == nil {
		itemsVal := reflect.ValueOf(items).Elem()
		for i := 0; i < itemsVal.Len(); i++ {
			var item2 interface{}
			item2 = itemsVal.Index(i).Interface()
			e.saveItemWithRoles(resource, item2, roles)
		}
	}

	return nil
}

func (e *adminSearch) saveItem(resource *Resource, item interface{}) error {
	roles := resource.Admin.getResourceViewRoles(*resource)
	return e.saveItemWithRoles(resource, item, roles)
}

func (e *adminSearch) saveItemWithRoles(resource *Resource, item interface{}, roles []string) error {
	relData := resource.itemToRelationData(item)
	if relData == nil {
		return errors.New("wrong item to relation data conversion")
	}
	searchItem := relationDataToSearchItem(resource, *relData)
	searchItem.Roles = roles
	return e.AddItem(&searchItem, 100)
}

func (e *adminSearch) deleteItem(resource *Resource, id int64) error {
	_, err := e.client.Delete().Index(e.indexName).Type(searchType).Id(searchID(resource, id)).Do(context.Background())
	return err
}

func searchID(resource *Resource, id int64) string {
	return fmt.Sprintf("%s-%d", resource.ID, id)
}

func relationDataToSearchItem(resource *Resource, data viewRelationData) SearchItem {
	return SearchItem{
		ID:          searchID(resource, data.ID),
		Category:    resource.HumanName("cs"),
		Name:        data.Name,
		Description: data.Description,
		Image:       data.Image,
		URL:         data.URL,
	}

}

func bindSearch(admin *Administration) {
	var err error

	adminSearch, err := NewAdminSearch(admin)
	if err != nil {
		admin.App.Log().Println(err)
		return
	} else {
		admin.search = adminSearch
	}

	go func() {
		err := adminSearch.searchImport()
		if err != nil {
			admin.App.Log().Println(fmt.Errorf("%s", err))
		}
	}()

	admin.AdminController.Get(admin.GetURL("_search"), func(request prago.Request) {
		q := request.Params().Get("q")
		pageStr := request.Params().Get("page")

		var page int = 1
		if pageStr != "" {
			var err error
			page, err = strconv.Atoi(pageStr)
			if err != nil {
				render404(request)
				return
			}
		}

		result, hits, err := adminSearch.Search(q, GetUser(request).GetRole(), page-1)
		must(err)

		var pages int = int(hits) / searchPageSize
		if hits > 0 {
			pages++
		}

		var searchPages []SearchPage
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
			searchPages = append(searchPages, SearchPage{
				Title:    i,
				Selected: selected,
				URL:      "_search?" + values.Encode(),
			})
		}

		title := fmt.Sprintf("Vyhledávání – \"%s\" – %d výsledků", q, hits)

		request.SetData("search_q", q)
		request.SetData("admin_title", title)
		request.SetData("search_results", result)
		request.SetData("search_pages", searchPages)

		request.SetData("admin_yield", "admin_search")
		request.RenderView("admin_layout")
	})

	/*mainController.Get("/suggest", func(request prago.Request) {
		results, err := elasticClient.Suggest(request.Params().Get("q"))
		if err != nil {
			request.Log().Println(err)
		}

		request.SetData("items", results)
		request.RenderView("suggest")
	})
	*/
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
