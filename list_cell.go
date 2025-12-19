package prago

import (
	"fmt"
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
	return &listCell{Name: crop(value.(string), 100), ItemID: f.id}
}

func markdownListDataSource(userData UserData, f *Field, value interface{}) *listCell {
	return &listCell{Name: cropMarkdown(value.(string), 100), ItemID: f.id}
}

func relationCellViewData(userData UserData, f *Field, value interface{}) *listCell {

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

func imageCellViewData(userData UserData, f *Field, value interface{}) *listCell {
	data := value.(string)
	ret := &listCell{
		ItemID: f.id,
	}
	ret.Images = append(ret.Images, data)
	return ret
}
