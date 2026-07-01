package prago

import (
	"fmt"
	"net/url"
	"strings"
)

type listCell struct {
	Images   []string
	Name     string
	ItemID   string
	FetchURL string
}

func (cell listCell) HasImages() bool {
	return len(cell.Images) > 0
}

func getCellViewData(userData UserData, field *Field, value any) *listCell {
	return field.fieldType.listCellDataSource(userData, field, value)
}

func basicCellDataSource(fn func(userData UserData, field *Field, value any) string) func(ud UserData, field *Field, item any) *listCell {
	return func(ud UserData, field *Field, item any) *listCell {
		return &listCell{
			Name:   fn(ud, field, item),
			ItemID: field.id,
		}
	}
}

func textListDataSource(userData UserData, f *Field, value any) *listCell {
	return &listCell{Name: value.(string), ItemID: f.id}
}

func markdownListDataSource(userData UserData, f *Field, value any) *listCell {
	return &listCell{Name: filterMarkdown(value.(string)), ItemID: f.id}
}

func relationCellViewData(userData UserData, f *Field, value any) *listCell {
	ret := &listCell{
		Name: "",
	}

	var urlData url.Values = map[string][]string{}
	urlData.Add("resource_id", f.resource.id)
	urlData.Add("field_id", f.id)

	var ids string
	intVal, ok := value.(int64)
	if ok {
		ids = fmt.Sprintf("%d", intVal)
	} else {
		ids = value.(string)
	}

	if ids == "" || ids == "0" {
		return &listCell{
			Name: "",
		}
	}

	urlData.Add("item_ids", ids)
	ret.FetchURL = "/admin/api/_fetch_list_cell_relation?" + urlData.Encode()

	return ret
}

func fetchListCellRelationAPIHandler(request *Request) {
	resource := request.app.resourceNameMap[request.Param("resource_id")]
	if !request.Authorize(resource.canView) {
		panic("not allowed resource")
	}

	field := resource.fieldMap[request.Param("field_id")]

	previewData := field.relationPreview(request, request.Param("item_ids"))

	var names []string
	var images []string

	if previewData != nil {
		for _, prev := range previewData {
			if prev.Image != "" {
				images = append(images, request.app.thumb(prev.ImageID))
			}
			names = append(names, prev.Name)
		}
	}

	request.WriteJSON(200, listCellFetchResponse{
		Name:   strings.Join(names, ", "),
		Images: images,
	})

}

func imageCellViewData(userData UserData, f *Field, value any) *listCell {
	data := value.(string)
	ret := &listCell{
		ItemID: f.id,
	}
	ret.Images = append(ret.Images, data)
	return ret
}
