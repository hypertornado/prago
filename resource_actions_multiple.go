package prago

import (
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
)

func (resource *Resource) initDefaultResourceMultipleActions() {

	resource.formItemMultipleAction(
		"edit-multiple",
		func(items []any, form *Form, request *Request) {

			var fieldOptions [][2]string
			for _, field := range resource.fields {
				if !request.Authorize(field.canEdit) {
					continue
				}
				fieldOptions = append(fieldOptions, [2]string{field.id, field.name(request.Locale())})
			}

			form.AddRadio("_field", "Jakou položku chcete opravit?", fieldOptions)

			var item any = reflect.New(resource.typ).Interface()
			var queryData url.Values = make(url.Values)
			for k, v := range resource.defaultValues {
				queryData.Set(k, v(request))
			}
			resource.bindData(item, request, queryData)
			form.initWithResourceItem(resource, item, request)

			for _, v := range form.Items {
				if v.ID == "_field" {
					continue
				}
				v.TextOver = v.Name
			}

			form.AddSubmit("Upravit položky")
		},
		func(items []any, fv FormValidation, request *Request) {
			field := resource.fieldMap[request.Param("_field")]
			if field == nil {
				fv.AddItemError("_field", "Vyberte jaké pole upravit")
			} else {
				if !request.Authorize(field.canEdit) {
					fv.AddItemError("_field", "Nelze opravit pole")
				}
			}
			if len(items) == 0 {
				fv.AddError("Vyberte položky")
			}

			if !fv.Valid() {
				return
			}

			for _, item := range items {
				val := reflect.ValueOf(item).Elem()
				id := val.FieldByName("ID").Int()
				params := make(url.Values)
				params.Set("id", fmt.Sprintf("%d", id))
				params.Set(field.id, request.Param(field.id))
				_, validation := resource.editItemWithLogAndValues(request, params)
				if !validation.valid {
					fv.AddError(validation.TextErrorReport(id, "cs").Text)
					return
				}
			}

			fv.Data(true)

		},
	).Icon(iconEdit).Permission(resource.canUpdate).Name(unlocalized("Upravit"))

	resource.formItemMultipleAction(
		"export-multiple",
		func(items []any, form *Form, request *Request) {
			form.AddSubmit("Export do .xlsx")
		},
		func(items []any, fv FormValidation, request *Request) {
			var ids []string
			for _, item := range items {
				val := reflect.ValueOf(item).Elem()
				id := val.FieldByName("ID").Int()
				ids = append(ids, fmt.Sprintf("%d", id))
			}
			redirectURL := fmt.Sprintf("/admin/%s/api/export?ids=%s", resource.id, strings.Join(ids, ","))
			fv.Redirect(redirectURL)
		},
	).Icon(iconDownload).Permission(resource.canExport).Name(unlocalized("Export"))

	resource.formItemMultipleAction(
		"ai-context",
		func(items []any, form *Form, request *Request) {
			for _, field := range resource.fields {
				if !request.Authorize(field.canView) {
					continue
				}
				checkboxItem := form.AddCheckbox(field.id, field.name(request.Locale()))
				checkboxItem.Value = "on"
			}

			for _, stat := range resource.itemStats {
				if !request.Authorize(stat.Permission) {
					continue
				}
				checkboxItem := form.AddCheckbox(stat.id, stat.Name(request.Locale()))
				checkboxItem.Value = "on"
			}

			if resource.previewFn != nil {
				form.AddCheckbox("_previewurl", "Ukázat preview URL").Value = "on"
			}

			form.AddHidden("_resource").Value = resource.id
			form.AddRelationMultiple("_items", resource.pluralName(request.Locale()), resource.id).Value = request.Param("_ids")

			form.AddSubmit("Zobrazit")
			//form.AddSubmit("AI Kontext")
		},
		func(items []any, fv FormValidation, request *Request) {
			var fields []*Field
			//resource := app.getResourceByID(request.Param("_resource"))
			for _, field := range resource.fields {
				if !request.Authorize(field.canView) {
					continue
				}
				if request.Param(field.id) == "on" {
					fields = append(fields, field)
				}
			}

			ids := MultirelationStringToArray(request.Param("_items"))
			if len(ids) == 0 {
				fv.AddItemError("_items", "Není vybrána žádná položka")
			}

			if !fv.Valid() {
				return
			}

			var strVal string

			for _, id := range ids {
				item := resource.query(context.Background()).ID(id)

				for _, field := range fields {
					ifaceVal := reflect.ValueOf(item).Elem().FieldByName(field.fieldClassName).Interface()

					cellData := getCellViewData(request, field, ifaceVal)
					strVal += fmt.Sprintf("%s: %v\n", field.name(request.Locale()), cellData.Name)
				}

				for _, stat := range resource.itemStats {
					if !request.Authorize(stat.Permission) {
						continue
					}

					if request.Param(stat.id) == "on" {
						strVal += fmt.Sprintf("%s: %v\n", stat.Name(request.Locale()), stat.Handler(item))
					}
				}

				if resource.previewFn != nil && request.Param("_previewurl") == "on" {
					prevURL := resource.previewFn(item)
					strVal += fmt.Sprintf("Veřejné URL: %v\n", prevURL)
				}

				strVal += "\n\n---------\n\n"
			}

			fv.AfterContent(template.HTML(fmt.Sprintf("<textarea class=\"input\">%s</textarea>", template.HTMLEscapeString(strVal))))
		},
	).Icon(iconDelete).Permission(resource.canDelete).Name(unlocalized("AI Kontext"))

	resource.formItemMultipleAction(
		"clone-multiple",
		func(items []any, form *Form, request *Request) {
			countEl := form.AddNumberInput("count", "Počet kopií")
			countEl.Value = "1"
			countEl.Focused = true

			form.AddSubmit("Naklonovat")
		},
		func(items []any, fv FormValidation, request *Request) {

			count, err := strconv.Atoi(request.Param("count"))
			if err != nil || count < 1 {
				fv.AddItemError("count", "Nesprávný počet")
			}

			if !fv.Valid() {
				return
			}

			for range count {
				for _, item := range items {
					val := reflect.ValueOf(item).Elem()
					val.FieldByName("ID").SetInt(0)
					timeVal := reflect.ValueOf(time.Now())
					for _, fieldName := range []string{"CreatedAt", "UpdatedAt"} {
						field := val.FieldByName(fieldName)
						if field.IsValid() && field.CanSet() && field.Type() == timeVal.Type() {
							field.Set(timeVal)
						}
					}

					validation := resource.validateUpdate(item, request)
					if !validation.Valid() {
						fv.AddError(fmt.Sprintf("Nelze naklonovat položku: %s", validation.TextErrorReport(0, request.Locale()).Text))
						return
					}

					err := resource.createWithLog(item, request)
					if err != nil {
						panic(fmt.Sprintf("can't create item for clone %v: %s", item, err))
					}

					must(
						resource.logActivity(request, nil, item),
					)

				}
			}
			fv.Data(true)
		},
	).Icon(iconDuplicate).Permission(resource.canCreate).Name(unlocalized("Naklonovat"))

	resource.formItemMultipleAction(
		"delete-multiple",
		func(items []any, form *Form, request *Request) {
			form.AddDeleteSubmit(messages.Get(request.Locale(), "admin_delete"))
		},
		func(items []any, fv FormValidation, request *Request) {
			for _, item := range items {
				vc := resource.validateDelete(item, request)
				for _, err := range vc.errors {
					fv.AddError(fmt.Sprintf("%s %s", err.Field, err.Text))
				}
			}
			if !fv.Valid() {
				return
			}
			for _, item := range items {
				must(resource.deleteWithLog(item, request))
			}
			request.AddFlashMessage(messages.Get(request.Locale(), "admin_item_deleted"))
			fv.Data(true)
		},
	).Icon(iconDelete).setPriority(-defaultHighPriority).styleDestroy().Permission(resource.canDelete).Name(messages.GetNameFunction("admin_delete"))

}
