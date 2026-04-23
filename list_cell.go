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

func getCellViewData(userData UserData, f *Field, value interface{}) *listCell {
	if f.fieldType.listCellDataSource != nil {
		return f.fieldType.listCellDataSource(userData, f, value)
	}

	if f.fieldType.isRelation() {
		return relationCellViewData(userData, f, value)
	}

	ret := &listCell{
		Name:   getDefaultFieldStringer(f)(userData, f, value),
		ItemID: f.id,
	}
	return ret
}

func textListDataSource(userData UserData, f *Field, value interface{}) *listCell {
	return &listCell{Name: value.(string), ItemID: f.id}
}

func markdownListDataSource(userData UserData, f *Field, value interface{}) *listCell {
	return &listCell{Name: filterMarkdown(value.(string)), ItemID: f.id}
}

func relationCellViewDataFetch(userData UserData, f *Field, value interface{}) *listCell {
	ret := &listCell{
		Name: "",
		//FetchURL: "/teest",
	}

	var urlData url.Values = map[string][]string{}
	urlData.Add("resource_id", f.resource.id)
	urlData.Add("field_id", f.id)
	//urlData.Add("item_id", stat.id)

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

func relationCellViewData(userData UserData, f *Field, value interface{}) *listCell {

	return relationCellViewDataFetch(userData, f, value)

	var ids string

	intVal, ok := value.(int64)
	if ok {
		ids = fmt.Sprintf("%d", intVal)
	} else {
		ids = value.(string)
	}

	previewData := f.relationPreview(userData, ids)
	if previewData == nil {
		return &listCell{}
	}

	var names []string
	var images []string
	for _, prev := range previewData {
		if prev.Image != "" {
			images = append(images, prev.ImageID)
		}
		names = append(names, prev.Name)
	}

	ret := &listCell{
		Name:   strings.Join(names, ", "),
		Images: images,
	}
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

func imageCellViewData(userData UserData, f *Field, value interface{}) *listCell {
	data := value.(string)
	ret := &listCell{
		ItemID: f.id,
	}
	ret.Images = append(ret.Images, data)
	return ret
}
