package prago

import "html/template"

func (timeline *Timeline) GetLegend(locale string) *timelineLegend {
	if len(timeline.dataSources) == 0 {
		return nil
	}
	if len(timeline.dataSources) == 1 && timeline.dataSources[0].name == nil {
		return nil
	}

	ret := &timelineLegend{}
	for _, v := range timeline.dataSources {
		item := &timelineLegendItem{
			Color: template.CSS(v.color),
		}
		if v.name != nil {
			item.Name = v.name(locale)
		}
		ret.Items = append(ret.Items, item)
	}
	return ret
}

type timelineLegend struct {
	Items []*timelineLegendItem
}

type timelineLegendItem struct {
	Color template.CSS
	Name  string
}
