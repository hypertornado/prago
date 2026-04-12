package prago

import "strings"

type TimelineDataFilter struct {
	KeyName   string
	ValueName string
}

func (timeline *Timeline) setDataFilter(data *timelineData, options map[string]string) {
	data.Filters = []*TimelineDataFilter{}
	for k, v := range options {
		if strings.HasPrefix(k, "_") {
			continue
		}
		if v == "" {
			continue
		}

		customFn := timeline.filterCustomNames[k]

		keyName := k
		valueName := v

		if customFn != nil {
			keyName, valueName = customFn(v)
		}

		data.Filters = append(data.Filters, &TimelineDataFilter{
			KeyName:   keyName,
			ValueName: valueName,
		})
	}
}
