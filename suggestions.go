package prago

import (
	"fmt"
	"reflect"
	"strconv"
)

type Suggestions struct {
	Message     string
	Suggestions []*Suggestion
	Button      *SuggestionsButton
}

type Suggestion struct {
	ID          int64
	Image       string
	URL         string
	Name        string
	Description string
}

type SuggestionsButton struct {
	Name    string
	FormURL string
}

func previewtoSuggestion(preview *Preview) *Suggestion {
	return &Suggestion{
		ID:          preview.ID,
		Image:       preview.Image,
		URL:         preview.URL,
		Name:        preview.Name,
		Description: preview.Description,
	}
}

func suggestionsResource(request *Request) {

	app := request.app

	resource := app.getResourceByID(request.Param("resource"))
	if resource == nil {
		panic("can't find resource")
	}
	if !request.Authorize(resource.canView) {
		panic("not allowed")

	}

	q := request.Param("q")

	filterID := request.Param("filter")
	//app := request.app

	var formFilter *FormFilter
	if filterID != "" {
		formFilter = app.formFilters[filterID]
		if formFilter == nil {
			panic(fmt.Sprintf("can't find filter id %s", filterID))
		}
	}

	ret := &Suggestions{}

	usedIDs := map[int64]bool{}

	previews := []*Preview{}

	id, err := strconv.Atoi(q)
	if err == nil {
		item := resource.query(request.r.Context()).ID(id)
		if item != nil {
			relationItem := resource.previewer(request, item).Preview(nil)
			if relationItem != nil {
				usedIDs[relationItem.ID] = true
				previews = append(previews, relationItem)
			}
		}
	}

	for _, fn := range resource.customSearchFunctions {
		previews = append(previews, fn(q, request)...)
	}

	filter := "%" + q + "%"

	var searchLimit int64 = 25

	searchableFields := resource.getSearchableFields(request)
	for _, field := range searchableFields {
		if field == nil {
			continue
		}

		query := resource.query(request.r.Context())
		if formFilter != nil {
			query = formFilter.filterFunction(query)
		}

		items, err := query.Limit(searchLimit).where(field.id+" LIKE ?", filter).OrderDesc("id").list()
		if err != nil {
			panic(err)
		}

		itemVals := reflect.ValueOf(items)
		itemLen := itemVals.Len()
		for i := 0; i < itemLen; i++ {
			viewItem := resource.previewer(request, itemVals.Index(i).Interface()).Preview(nil)
			if viewItem != nil && !usedIDs[viewItem.ID] {
				usedIDs[viewItem.ID] = true
				previews = append(previews, viewItem)
			}
		}
	}

	if len(previews) > int(searchLimit) {
		previews = previews[0:searchLimit]
	}

	if (len(previews)) == 0 {
		ret.Message = "Nic nenalezeno"
	}

	ret.Suggestions = []*Suggestion{}

	for _, v := range previews {
		ret.Suggestions = append(ret.Suggestions, previewtoSuggestion(v))
	}

	if request.Authorize(resource.canCreate) {
		ret.Button = &SuggestionsButton{
			Name:    resource.newItemName(request.Locale()),
			FormURL: resource.getURL("new"),
		}
	}

	request.WriteJSON(200, ret)
}
