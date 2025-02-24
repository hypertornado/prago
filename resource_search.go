package prago

import (
	"reflect"
	"strconv"
)

type searchResourceResponse struct {
	Message  string
	Previews []preview
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

	previews := []preview{}

	id, err := strconv.Atoi(q)
	if err == nil {
		item := resource.query(request.r.Context()).ID(id)
		if item != nil {
			relationItem := resource.previewer(request, item).Preview(nil)
			if relationItem != nil {
				usedIDs[relationItem.ID] = true
				previews = append(previews, *relationItem)
			}
		}
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
				previews = append(previews, *viewItem)
			}
		}
	}

	if len(previews) > 5 {
		previews = previews[0:5]
	}

	/*for k := range previews {
		//TODO: remove this crop
		previews[k].Description = crop(previews[k].Description, 100)
	}*/

	if (len(previews)) == 0 {
		ret.Message = "Nic nenalezeno"
	}

	ret.Previews = previews

	if request.Authorize(resource.canCreate) {

		ret.Button = &searchResourceResponseButton{
			Name:    messages.GetNameFunction("admin_new")(request.Locale()) + " - " + resource.singularName(request.Locale()),
			FormURL: resource.getURL("new"),
		}
	}

	request.WriteJSON(200, ret)
}
