package prago

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type SuggestionsRequest struct {
	CustomUUID    string
	Resource      string
	Q             string
	Filter        string
	ContextValues map[string]string
}

type SuggestionsResponse struct {
	Message     string
	Suggestions []*Suggestion
	Buttons     []*Button
}

type Suggestion struct {
	ID          string
	Image       string
	URL         string
	Name        string
	Description string

	SetValues map[string]string
}

func (app *App) initSuggestions() {
	app.initSuggestionsAdvanced()
	app.initSuggestionsCustom()

	app.NewAPI("_suggestions").Handler(suggestionsResource).Permission(loggedPermission).Method("POST")
}

func previewtoSuggestion(preview *Preview) *Suggestion {
	return &Suggestion{
		ID:          fmt.Sprintf("%d", preview.ID),
		Image:       preview.Image,
		URL:         preview.URL,
		Name:        preview.Name,
		Description: preview.Description,
	}
}

func suggestionsResource(request *Request) {

	data, err := io.ReadAll(request.Request().Body)
	Must(err)

	var suggestionRequest SuggestionsRequest
	err = json.Unmarshal(data, &suggestionRequest)
	Must(err)

	var ret *SuggestionsResponse

	if suggestionRequest.CustomUUID != "" {
		ret = suggestionRequest.getSuggestionsCustom(request)
	} else {
		ret = suggestionRequest.getSuggestionsResource(request)
	}
	request.WriteJSON(200, ret)
}

func (suggestionRequest SuggestionsRequest) getSuggestionsCustom(request *Request) *SuggestionsResponse {
	suggestion := request.app.getSuggestionCustom(suggestionRequest.CustomUUID)
	if suggestion == nil {
		return nil
	}
	if !request.Authorize(suggestion.permission) {
		return nil
	}
	return suggestion.handler(suggestionRequest)
}

func (suggestionRequest SuggestionsRequest) getSuggestionsResource(request *Request) *SuggestionsResponse {
	app := request.app
	resource := app.getResourceByID(suggestionRequest.Resource)
	if resource == nil {
		panic("can't find resource")
	}
	if !request.Authorize(resource.canView) {
		panic("not allowed")

	}

	q := suggestionRequest.Q
	filterID := suggestionRequest.Filter

	var formFilter *FormFilter
	if filterID != "" {
		formFilter = app.formFilters[filterID]
		if formFilter == nil {
			panic(fmt.Sprintf("can't find filter id %s", filterID))
		}
	}

	ret := &SuggestionsResponse{}

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
		ret.Buttons = append(ret.Buttons, &Button{
			Icon: "glyphicons-basic-371-plus.svg",
			Name: resource.newItemName(request.Locale()),
			URL:  resource.getURL("new"),
		})
	}

	ret.Buttons = append(ret.Buttons, &Button{
		Icon: "glyphicons-basic-395-filter.svg",
		Name: fmt.Sprintf("Filtrovat „%s“", resource.pluralName("cs")),
		URL:  fmt.Sprintf("/admin/_suggestionsadvanced?resource=%s", resource.id),
	})

	return ret

}
