package administration

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hypertornado/prago"
	"github.com/olivere/elastic"
	"golang.org/x/net/context"
)

type SearchItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
	URL         string `json:"url"`
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
            "url": {"type": "text"}
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
	_, err := e.client.Index().Index(e.indexName).Type("items").BodyJson(map[string]interface{}{
		"suggest": map[string]interface{}{
			"input":  suggest,
			"weight": weight,
		},
		"name":        item.Name,
		"description": item.Description,
		"image":       item.Image,
		"url":         item.URL,
	}).Id(item.ID).Do(context.Background())
	return err
}

func (e *adminSearch) DeleteIndex() error {
	_, err := e.client.DeleteIndex(e.indexName).Do(context.Background())
	return err
}

func (e *adminSearch) Search(q string, limit int) ([]*SearchItem, error) {
	var ret []*SearchItem

	mq := elastic.NewMultiMatchQuery(q)
	mq.FieldWithBoost("name", 3)
	mq.FieldWithBoost("description", 2)
	mq.FieldWithBoost("text", 1)

	searchResult, err := e.client.Search().
		Index(e.indexName).
		Type("items").
		Query(mq).
		From(0).
		Size(limit).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	var item SearchItem
	for _, item := range searchResult.Each(reflect.TypeOf(item)) {
		if t, ok := item.(SearchItem); ok {
			ret = append(ret, &t)
		}
	}

	return ret, nil
}

func (e *adminSearch) Suggest(q string) ([]*SearchItem, error) {
	suggesterName := "completion_suggester"
	cs := elastic.NewCompletionSuggester(suggesterName)
	cs = cs.Field("suggest")
	cs = cs.Prefix(q)
	cs = cs.SkipDuplicates(true)

	searchResult, err := e.client.Search().
		Index(e.indexName).
		Type("items").
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
	//searchImportMutex.Lock()
	//defer searchImportMutex.Unlock()
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

	/*
		var categories []*PackageCategory
		must(admin.Query().WhereIs("language", "cs").WhereIs("hidden", false).Get(&categories))
		i := 0
		for _, v := range categories {
			err := client.AddItem(&SearchItem{
				ID:          int64(i),
				Typ:         "package-list",
				Name:        v.Name,
				Description: v.Description,
				Url:         v.getNavTab("", "").URL,
			}, 10)
			i++
			if err != nil {
				return err
			}
		}*/

	//lastSearchImport = time.Now()

	return nil
}

//func importResource()

/*func SearchCronTask() {
	if time.Now().Add(-23 * time.Hour).Before(lastSearchImport) {
		if time.Now().Hour() == 3 {
			searchImport()
		}
	}
}*/

func (e *adminSearch) importResource(resource *Resource) error {
	fmt.Printf("importing %s\n", resource.HumanName("cs"))

	var item interface{}
	resource.newItem(&item)
	c, _ := e.admin.Query().Count(item)

	if c > 100 {
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
			relData := resource.itemToRelationData(item2)
			if relData == nil {
				continue
			}
			searchItem := relationDataToSearchItem(resource, *relData)
			e.AddItem(&searchItem, 100)
			//err := e.AddItem(&searchItem, 100)
			//must(err)
			if false {
				fmt.Println(searchItem)
			}
		}
	}

	return nil
}

func relationDataToSearchItem(resource *Resource, data viewRelationData) SearchItem {
	id := fmt.Sprintf("%s-%d", resource.ID, data.ID)
	return SearchItem{
		ID:          id,
		Name:        data.Name,
		Description: data.Description,
		Image:       data.Image,
		URL:         data.URL,
	}

}

func bindSearch(admin *Administration) {
	var err error

	adminSearch, err := NewAdminSearch(admin)

	//elasticClient, err = NewClient(elasticsearchIndexName)
	if err != nil {
		admin.App.Log().Println(err)
		return
	}

	/*admin.App.AddCronTask("index search", SearchCronTask, func(in time.Time) time.Time {
		return in.Add(30 * time.Minute)
	})

	admin.App.AddCommand("indexsearch").Callback(SearchCronTask)*/

	go func() {
		err := adminSearch.searchImport()
		if err != nil {
			admin.App.Log().Println(fmt.Errorf("%s", err))
		}
	}()

	admin.AdminController.Get(admin.GetURL("_search"), func(request prago.Request) {
		q := request.Params().Get("q")

		result, err := adminSearch.Search(q, 10)
		must(err)

		request.SetData("search_q", q)
		request.SetData("admin_title", q)
		request.SetData("search_results", result)

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

	mainController.Get("/hledej", func(request prago.Request) {
		q := request.Params().Get("q")
		results, err := elasticClient.Search(q, 100)
		if err != nil {
			panic(err)
		}

		request.SetData("name", "Vyhledávání "+q)
		request.SetData("description", fmt.Sprintf("Počet výsledků: %d", len(results)))
		request.SetData("content_after", "search")
		request.SetData("q", q)

		request.SetData("Items", results)
		request.SetData("content_after", "list")

		request.SetData("dont_show_endorsement", true)
		pageYield(request)
	})*/
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
