package prago

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"

	"github.com/hypertornado/prago/pragelastic"
	"github.com/olivere/elastic/v7"
)

type searchItem struct {
	ID              string
	SuggestionField pragelastic.Suggest `elastic-analyzer:"czech_suggest" elastic-category-context-name:"Roles"`
	Category        string              `elastic-datatype:"keyword"`
	Name            string              `elastic-datatype:"text" elastic-analyzer:"czech"`
	Description     string              `elastic-datatype:"text" elastic-analyzer:"czech"`
	Image           string              `elastic-datatype:"keyword"`
	URL             string              `elastic-datatype:"keyword"`
	Roles           []string
}

const searchPageSize int64 = 10

type searchPage struct {
	Title    int
	Selected bool
	URL      string
}

type adminSearch struct {
	app   *App
	index *pragelastic.Index[searchItem]
}

func (app *App) initElasticsearchClient() {
	client, err := pragelastic.New(app.codeName)
	if err != nil {
		app.Log().Printf("initElasticsearchClient, client can't be initiated: %s", err)
	}
	app.ElasticClient = client
}

func (app *App) initSearch() {
	if app.ElasticClient == nil {
		app.Log().Println("will not initialize search since elasticsearch client is not defined")
		return
	}
	adminSearch, err := newAdminSearch(app)
	if err != nil {
		app.Log().Println("admin search not initialized: " + err.Error())
		return
	}
	app.Log().Println("admin search initialized")
	app.search = adminSearch

	app.sysadminTaskGroup.Task(unlocalized("index_search")).Handler(func(ta *TaskActivity) error {
		return adminSearch.searchImport()
	})

	app.Action("_search").Permission(loggedPermission).Name(unlocalized("Vyhledávání")).Template("admin_search").hiddenInMainMenu().DataSource(
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

			result, hits, err := adminSearch.Search(q, request.user.Role, int64(page-1), request.Request().Context())
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
			results, err := adminSearch.Suggest(request.Param("q"), request.user.Role)
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

func newAdminSearch(app *App) (*adminSearch, error) {
	index := pragelastic.NewIndex[searchItem](app.ElasticClient)

	return &adminSearch{
		app:   app,
		index: index,
	}, nil
}

func (resourceData *resourceData) importSearchData(bulkUpdater *pragelastic.BulkUpdater[searchItem]) error {
	roles := resourceData.getResourceViewRoles()
	name := resourceData.pluralName("cs")
	var resourceSearchItem = searchItem{
		ID: "resource_" + resourceData.id,
		SuggestionField: pragelastic.Suggest{
			Input:  name,
			Weight: 100,
			Contexts: map[string][]string{
				"Roles": roles,
			},
		},
		Name:  name,
		URL:   resourceData.getURL(""),
		Roles: roles,
	}

	resourceData.app.search.addItem(bulkUpdater, &resourceSearchItem, 200)

	c, _ := resourceData.query().count()
	if c > 10000 {
		return nil
	}

	items, err := resourceData.query().list()
	if err != nil {
		return err
	}

	itemVals := reflect.ValueOf(items)
	itemLen := itemVals.Len()
	for i := 0; i < itemLen; i++ {
		resourceData.saveSearchItemWithRoles(bulkUpdater, itemVals.Index(i).Interface(), roles)
	}

	/*for _, item := range items {
		resource.saveSearchItemWithRoles(bulkUpdater, item, roles)
	}*/
	return nil
}

func (e *adminSearch) flush() error {
	return e.index.Flush()
}

func (resourceData *resourceData) saveSearchItem(item any) error {
	roles := resourceData.getResourceViewRoles()
	return resourceData.saveSearchItemWithRoles(nil, item, roles)
}

func (resourceData *resourceData) saveSearchItemWithRoles(bulkUpdater *pragelastic.BulkUpdater[searchItem], item any, roles []string) error {
	//TODO: ugly hack
	preview := resourceData.getPreview(item, &user{}, nil)
	if preview == nil {
		return errors.New("wrong item to relation data conversion")
	}
	searchItem := relationDataToSearchItem(resourceData, *preview, roles)
	searchItem.Roles = roles
	return resourceData.app.search.addItem(bulkUpdater, &searchItem, 100)
}

func relationDataToSearchItem(resourceData *resourceData, data preview, roles []string) searchItem {
	return searchItem{
		ID: searchID(resourceData, data.ID),
		SuggestionField: pragelastic.Suggest{
			Input:  data.Name,
			Weight: 10,
			Contexts: map[string][]string{
				"Roles": roles,
			},
		},
		Category:    resourceData.pluralName("cs"),
		Name:        data.Name,
		Description: data.Description,
		Image:       data.Image,
		URL:         data.URL,
	}
}

func searchID(resourceData *resourceData, id int64) string {
	return fmt.Sprintf("%s-%d", resourceData.getID(), id)
}

func (e *adminSearch) deleteItem(resourceData *resourceData, id int64) error {
	if e.index != nil {
		return e.index.DeleteItem(searchID(resourceData, id))
	}
	return nil
}

func (e *adminSearch) searchImport() error {
	var err error

	e.app.Log().Println("Importing admin search index")

	bulkUpdater, err := e.index.UpdateBulk()
	if err != nil {
		return fmt.Errorf("can't create bulk update: %s", err)
	}

	err = e.createSearchIndex()
	if err != nil {
		return fmt.Errorf("while creating index: %s", err)
	}

	for _, resourceData := range e.app.resources {
		total, _ := e.index.Count()
		e.app.Log().Printf("importing resource: %s, total: %d\n", resourceData.getID(), total)
		err = resourceData.importSearchData(bulkUpdater)
		if err != nil {
			return fmt.Errorf("while importing resource %s: %s", resourceData.getID(), err)
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
	e.app.Log().Println("INDEX Created")
	return nil
}

func (e *adminSearch) createSearchIndex() error {
	e.index.Delete()
	err := e.index.Create()
	if err != nil {
		return err
	}

	return nil
}

func (e *adminSearch) addItem(bulkUpdater *pragelastic.BulkUpdater[searchItem], item *searchItem, weight int) error {
	if bulkUpdater != nil {
		bulkUpdater.AddItem(item)
		return nil
	}
	if e.index != nil {
		return e.index.UpdateSingle(item)
	}
	return nil
}

func (e *adminSearch) DeleteIndex() error {
	return e.index.Delete()
}

func (e *adminSearch) Search(q string, role string, page int64, ctx context.Context) ([]*searchItem, int64, error) {

	mq := elastic.NewMultiMatchQuery(q)
	mq.FieldWithBoost("Name", 3)
	mq.FieldWithBoost("description", 1)

	items, totalHits, err := e.index.Query().
		Offset(page*searchPageSize).
		Limit(searchPageSize).
		ShouldQuery(mq).
		Filter("Roles", role).
		Context(ctx).
		List()
	if err != nil {
		return nil, -1, err
	}
	return items, totalHits, nil
}

func (e *adminSearch) Suggest(q string, role string) ([]*searchItem, error) {
	if role == "" {
		return nil, nil
	}
	return e.index.Suggest(q, map[string][]string{
		"Roles": {
			role,
		},
	})
}

func (si searchItem) CroppedDescription() string {
	return crop(si.Description, 100)
}
