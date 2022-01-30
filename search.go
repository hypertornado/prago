package prago

/*
const searchPageSize int = 10

type searchPage struct {
	Title    int
	Selected bool
	URL      string
}

type adminSearchOLD struct {
	client    *elastic.Client
	app       *App
	indexName string
}

func (si searchItem) CroppedDescription() string {
	return crop(si.Description, 100)
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

	return ret, nil
}

func (e *adminSearch) flush() error {
	_, err := e.client.Flush().Do(context.Background())
	return err
}

func (resource *Resource[T]) importSearchData() error {
	roles := resource.getResourceViewRoles()
	var resourceSearchItem = searchItem{
		ID:    "resource_" + resource.id,
		Name:  resource.getPluralNameFunction()("cs"),
		URL:   resource.getURL(""),
		Roles: roles,
	}
	resource.app.search.addItem(&resourceSearchItem, 200)

	c, _ := resource.Query().Count()
	if c > 10000 {
		return nil
	}

	items := resource.Query().List()
	for _, item := range items {
		resource.saveSearchItemWithRoles(item, roles)
	}

	return nil
}

func (resource *Resource[T]) saveSearchItem(item *T) error {
	roles := resource.getResourceViewRoles()
	return resource.saveSearchItemWithRoles(item, roles)
}

func (resource *Resource[T]) saveSearchItemWithRoles(item *T, roles []string) error {
	//TODO: ugly hack
	preview := resource.getPreview(item, &user{}, nil)
	if preview == nil {
		return errors.New("wrong item to relation data conversion")
	}
	searchItem := relationDataToSearchItem(resource, *preview)
	searchItem.Roles = roles
	return resource.app.search.addItem(&searchItem, 100)
}

func (e *adminSearch) deleteItem(resource resourceIface, id int64) error {
	_, err := e.client.Delete().Index(e.indexName).Id(searchID(resource, id)).Do(context.Background())
	return err
}

func searchID(resource resourceIface, id int64) string {
	return fmt.Sprintf("%s-%d", resource.getID(), id)
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
*/
