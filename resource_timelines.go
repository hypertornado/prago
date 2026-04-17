package prago

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func (resource *Resource) initResourceTimelines() {

	timelinesDashboard := resource.resourceBoard.Dashboard(unlocalized(""))

	timeTyp := reflect.TypeOf(time.Now())

	for _, field := range resource.fields {
		if field.typ != timeTyp {
			continue
		}

		timeline := timelinesDashboard.Timeline(field.name, field.canView, func(tdr *TimelineDataRequest) float64 {
			q := resource.query(tdr.Request.r.Context())

			for _, field2 := range resource.fields {
				if !tdr.Request.Authorize(field2.canView) {
					continue
				}
				if field2.fieldType.isRelation() {
					q.In(field2.id, tdr.Options[field2.id])
				}
			}

			ret, _ := q.
				where(fmt.Sprintf("`%s` >= ? AND `%s` < ?", field.id, field.id), tdr.From, tdr.To).
				count()
			return float64(ret)
		})

		timeline.OptionsForm(func(request *Request, form *Form) {
			for _, field2 := range resource.fields {
				if !request.Authorize(field2.canView) {
					continue
				}
				if field2.fieldType.isRelation() {
					form.AddRelationMultiple(field2.id, field2.name(request.Locale()), field2.getRelatedID()).Value = request.Param(field2.id)
				}
			}
		})

		for _, field2 := range resource.fields {
			if !field2.fieldType.isRelation() {
				continue
			}
			timeline.FilterName(field2.id, func(request *Request, value string) (string, string) {
				if !request.Authorize(field2.canEdit) {
					return field2.id, value
				}

				var names []string
				previews := field2.relationPreview(request, value)
				for _, v := range previews {
					names = append(names, v.Name)
				}

				return field2.name(request.Locale()), strings.Join(names, " · ")
			})

		}
	}

	if resource.activityLog {
		timelinesDashboard.Timeline(unlocalized("Úpravy"), resource.canView, func(tdr *TimelineDataRequest) float64 {

			q := resource.app.activityLogResource.query(tdr.Request.r.Context())
			q.Is("resourcename", resource.id)
			count, _ := q.where("`createdat` >= ? AND `createdat` < ?", tdr.From, tdr.To).count()
			return float64(count)
		})

	}
}
