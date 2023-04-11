package prago

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

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
	app.API("search-suggest").Permission(loggedPermission).Handler(
		func(request *Request) {
			results, err := app.suggestItems(request.Param("q"), request)
			must(err)
			if len(results) == 0 {
				request.RenderJSONWithCode(nil, 204)
				return
			}

			request.SetData("items", results)
			request.Write(200, "admin_search_suggest", request.data)
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

			request.SetData("search_q", q)

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
		return adminSearch.searchImport(context.TODO())
	})

}

func (app *App) searchItems(q string, page int64, request *Request) (ret []*searchItem, hits int64, err error) {
	if app.search != nil {
		ret, hits, err = app.search.Search(q, request.role(), page, request.r.Context())
		must(err)
		return
	} else {
		ret = app.searchWithoutElastic(q, request)
		hits = int64(len(ret))
	}
	return
}

func (app *App) suggestItems(q string, request *Request) (ret []*searchItem, err error) {
	q = strings.Trim(q, " ")
	if q == "" {
		return ret, nil
	}

	ret = app.searchWithoutElastic(q, request)

	if app.search != nil {
		elasticResults, _, err := app.searchItems(q, 0, request)
		if err != nil {
			app.Log().Println(err)
		} else {
			ret = append(ret, elasticResults...)
		}
	}

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

func newAdminSearch(app *App) (*adminSearch, error) {
	index := pragelastic.NewIndex[searchItem](app.ElasticClient)

	return &adminSearch{
		app:   app,
		index: index,
	}, nil
}

func (resourceData *resourceData) importSearchData(ctx context.Context, bulkUpdater *pragelastic.BulkUpdater[searchItem]) error {
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

	c, _ := resourceData.query(ctx).count()
	if c > 10000 {
		return nil
	}

	items, err := resourceData.query(ctx).list()
	if err != nil {
		return err
	}

	itemVals := reflect.ValueOf(items)
	itemLen := itemVals.Len()
	for i := 0; i < itemLen; i++ {
		resourceData.saveSearchItemWithRoles(ctx, bulkUpdater, itemVals.Index(i).Interface(), roles)
	}

	return nil
}

func (e *adminSearch) flush() error {
	return e.index.Flush()
}

func (resourceData *resourceData) saveSearchItem(ctx context.Context, item any) error {
	roles := resourceData.getResourceViewRoles()
	return resourceData.saveSearchItemWithRoles(context.TODO(), nil, item, roles)
}

func (resourceData *resourceData) saveSearchItemWithRoles(ctx context.Context, bulkUpdater *pragelastic.BulkUpdater[searchItem], item any, roles []string) error {

	//TODO: ugly hack with sysadmin user, remove suggestions
	previewer := resourceData.previewer(resourceData.app.newUserData(&user{Role: "sysadmin"}), item)
	if previewer == nil {
		return errors.New("wrong item to relation data conversion")
	}
	searchItem := relationDataToSearchItem(ctx, resourceData, previewer, roles)
	searchItem.Roles = roles
	return resourceData.app.search.addItem(bulkUpdater, &searchItem, 100)
}

func relationDataToSearchItem(ctx context.Context, resourceData *resourceData, previewer *previewer, roles []string) searchItem {
	return searchItem{
		ID: searchID(resourceData, previewer.ID()),
		SuggestionField: pragelastic.Suggest{
			Input:  previewer.Name(),
			Weight: 10,
			Contexts: map[string][]string{
				"Roles": roles,
			},
		},
		Category:    resourceData.pluralName("cs"),
		Name:        previewer.Name(),
		Description: previewer.DescriptionBasic(nil),
		Image:       previewer.ThumbnailURL(ctx),
		URL:         previewer.URL(""),
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

func (e *adminSearch) searchImport(ctx context.Context) error {
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
		err = resourceData.importSearchData(ctx, bulkUpdater)
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
	//e.app.Log().Println("INDEX Created")
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
