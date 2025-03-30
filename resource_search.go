package prago

import (
	"reflect"
	"strconv"
)

type searchResourceResponse struct {
	Message  string
	Previews []*Preview
	Button   *searchResourceResponseButton
}

type searchResourceResponseButton struct {
	Name    string
	FormURL string
}

func searchResource(request *Request, resource *Resource) {
	q := request.Param("q")

	ret := &searchResourceResponse{}

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

	searchableFields := resource.getSearchableFields(request)
	for _, field := range searchableFields {
		if field == nil {
			continue
		}
		items, err := resource.query(request.r.Context()).Limit(5).where(field.id+" LIKE ?", filter).OrderDesc("id").list()
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

	if len(previews) > 10 {
		previews = previews[0:10]
	}

	if (len(previews)) == 0 {
		ret.Message = "Nic nenalezeno"
	}

	ret.Previews = previews

	if request.Authorize(resource.canCreate) {
		ret.Button = &searchResourceResponseButton{
			Name:    resource.newItemName(request.Locale()),
			FormURL: resource.getURL("new"),
		}
	}

	request.WriteJSON(200, ret)
}

func AddResourceCustomSearchFunction[T any](app *App, fn func(q string, userData UserData) []*T) {
	resource := getResource[T](app)
	resource.customSearchFunctions = append(resource.customSearchFunctions,
		func(q string, userData UserData) (ret []*Preview) {
			items := fn(q, userData)
			for _, item := range items {
				preview := resource.previewer(userData, item).Preview(nil)
				ret = append(ret, preview)
			}
			return ret
		},
	)
}
