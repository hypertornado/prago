package prago

import (
	"context"
	"strings"
)

type listCell struct {
	Images []string
	Name   string
}

func getCellViewData(userData UserData, f *Field, value interface{}) listCell {
	if f.fieldType.listCellDataSource != nil {
		return f.fieldType.listCellDataSource(userData, f, value)
	}

	if f.fieldType.IsRelation() {
		return relationCellViewData(userData, f, value)
	}

	ret := listCell{
		Name: getDefaultFieldStringer(f)(userData, f, value),
	}
	return ret
}

func textListDataSource(userData UserData, f *Field, value interface{}) listCell {
	return listCell{Name: crop(value.(string), 100)}
}

func markdownListDataSource(userData UserData, f *Field, value interface{}) listCell {
	return listCell{Name: cropMarkdown(value.(string), 100)}
}

func relationCellViewData(userData UserData, f *Field, value interface{}) listCell {
	previewData := f.relationPreview(context.TODO(), userData, value.(int64))
	if previewData == nil {
		return listCell{}
	}

	ret := listCell{
		Name: previewData.Name,
	}
	if previewData.Image != "" {
		ret.Images = []string{previewData.Image}
	}
	return ret
}

func imageCellViewData(userData UserData, f *Field, value interface{}) listCell {
	data := value.(string)
	images := strings.Split(data, ",")
	ret := listCell{}
	if len(images) > 0 {
		ret.Images = images
	}
	return ret
}
