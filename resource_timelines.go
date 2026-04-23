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

				if field2.filterLayout() == "filter_layout_boolean" {
					val := tdr.Options[field2.id]
					if val == "true" {
						q.Is(field2.id, true)
					}
					if val == "false" {
						q.Is(field2.id, false)
					}
				}

				if field2.filterLayout() == "filter_layout_select" {
					val := tdr.Options[field2.id]
					if val != "" {
						q.Is(field2.id, val)
					}

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

				if field2.filterLayout() == "filter_layout_boolean" {
					form.AddSelect(field2.id, field2.name(request.Locale()), [][2]string{
						{"", ""},
						{"true", "✅ ano"},
						{"false", "ne"},
					}).Value = request.Param(field2.id)
				}

				if field2.filterLayout() == "filter_layout_select" {
					options := field2.fieldType.filterLayoutDataSource(field2, request).([][2]string)
					form.AddSelect(field2.id, field2.name(request.Locale()), options).Value = request.Param(field2.id)

				}
			}
		})

		for _, field2 := range resource.fields {
			if field2.fieldType.isRelation() {
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

			if field2.filterLayout() == "filter_layout_boolean" {
				timeline.FilterName(field2.id, func(request *Request, value string) (string, string) {
					name := value
					if value == "true" {
						name = "✅ ano"
					}
					if value == "false" {
						name = "ne"
					}
					return field2.name(request.Locale()), name
				})
			}

			if field2.filterLayout() == "filter_layout_select" {
				timeline.FilterName(field2.id, func(request *Request, value string) (string, string) {
					name := value
					if value != "" {
						name = field2.fieldType.viewDataSource(request, field2, value).(string)
					}
					return field2.name(request.Locale()), name
				})
			}

		}
	}

	if resource.activityLog {
		timeline := timelinesDashboard.Timeline(unlocalized("Úpravy"), resource.canView, func(tdr *TimelineDataRequest) float64 {

			q := resource.app.activityLogResource.query(tdr.Request.r.Context())
			q.Is("resourcename", resource.id)
			q.In("user", tdr.Options["users"])
			count, _ := q.where("`createdat` >= ? AND `createdat` < ?", tdr.From, tdr.To).count()
			return float64(count)
		})

		timeline.OptionsForm(func(request *Request, form *Form) {
			form.AddRelationMultiple("users", "Uživatel", "user").Value = request.Param("users")
		})

		timeline.FilterName("users", func(request *Request, value string) (string, string) {

			var names []string
			users := Query[user](request.app).In("id", value).List()
			for _, v := range users {
				names = append(names, v.Name)
			}

			return "Uživatel", strings.Join(names, " · ")
		})

	}
}
