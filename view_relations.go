package prago

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/net/context"
)

type viewRelation struct {
	SourceResource string
	TargetResource string
	TargetField    string
	IDValue        int64
	Count          int64
}

func (resource *Resource) getRelationViews(id int64, request *Request) (ret []*view) {
	for _, v := range resource.relations {
		vi := resource.getRelationView(id, v, request)
		if vi != nil {
			ret = append(ret, vi)
		}
	}
	return
}

func (resource *Resource) getRelationView(id int64, field *relatedField, request *Request) *view {
	if !request.Authorize(field.resource.canView) {
		return nil
	}

	filteredCount := field.resource.itemWithRelationCount(request.r.Context(), field.id, int64(id))

	ret := &view{}

	icon := iconResource
	if field.resource.icon != "" {
		icon = field.resource.icon
	}
	ret.Icon = icon

	name := field.listName(request.Locale())
	ret.Name = name
	ret.Subname = messages.ItemsCount(filteredCount, request.Locale())

	ret.Navigation = append(ret.Navigation, viewButton{
		Name: messages.Get(request.Locale(), "admin_list"),
		Icon: iconTable,
		URL:  field.listURL(int64(id)),
	})

	if request.Authorize(field.resource.canUpdate) {
		ret.Navigation = append(ret.Navigation, viewButton{
			Name: messages.Get(request.Locale(), "admin_new"),
			Icon: iconAdd,
			URL:  field.addURL(int64(id)),
		})
	}

	ret.Relation = &viewRelation{
		SourceResource: resource.getID(),
		TargetResource: field.resource.getID(),
		TargetField:    field.id,
		IDValue:        int64(id),
		Count:          filteredCount,
	}
	return ret
}

func (resource *Resource) itemWithRelationCount(ctx context.Context, fieldID string, id int64) int64 {

	field := resource.fieldMap[fieldID]
	if field.typ.Kind() == reflect.String {
		filteredCount, err := resource.query(ctx).where(fmt.Sprintf("%s LIKE '%%;%d;%%'", fieldID, id)).count()
		if err != nil {
			panic(err)
		}
		return filteredCount
	} else {
		filteredCount, err := resource.query(ctx).Is(fieldID, id).count()
		if err != nil {
			panic(err)
		}
		return filteredCount
	}
}

type relationListRequest struct {
	SourceResource string
	TargetResource string
	TargetField    string
	IDValue        int64
	Offset         int64
	Count          int64
}

func generateRelationListAPIHandler(request *Request) {
	decoder := json.NewDecoder(request.Request().Body)
	var listRequest relationListRequest
	decoder.Decode(&listRequest)
	defer request.Request().Body.Close()

	targetResource := request.app.getResourceByID(listRequest.TargetResource)
	data := targetResource.getPreviews(request.r.Context(), listRequest, request)

	request.WriteHTML(200, request.app.adminTemplates, "view_relationlist_response", data)
}

func (resource *Resource) getPreviews(ctx context.Context, listRequest relationListRequest, request *Request) []*Preview {
	sourceResource := resource.app.getResourceByID(listRequest.SourceResource)
	if !request.Authorize(sourceResource.canView) {
		panic("cant authorize source resource")
	}

	if !request.Authorize(resource.canView) {
		panic("cant authorize target resource")
	}

	q := resource.query(ctx)

	field := resource.fieldMap[listRequest.TargetField]
	if field.typ.Kind() == reflect.String {
		q.where(fmt.Sprintf("%s LIKE '%%;%d;%%'", listRequest.TargetField, listRequest.IDValue))
	} else {
		q.Is(listRequest.TargetField, fmt.Sprintf("%d", listRequest.IDValue))
	}

	if resource.orderDesc {
		q.addOrder(resource.orderByColumn, true)
	} else {
		q.addOrder(resource.orderByColumn, false)
	}

	limit := listRequest.Count
	if limit > 10 {
		limit = 10
	}
	q = q.Limit(limit)
	q.Offset(listRequest.Offset)

	rowItems, err := q.list()
	must(err)

	itemVals := reflect.ValueOf(rowItems)

	itemLen := itemVals.Len()

	var ret []*Preview

	for i := 0; i < itemLen; i++ {
		ret = append(
			ret,
			resource.previewer(request, itemVals.Index(i).Interface()).Preview(sourceResource),
		)
	}

	return ret
}

func humanizeMultiRelationsString(in string) string {
	var outFields []string
	fields := strings.Split(in, ";")
	for _, field := range fields {
		if field != "" {
			outFields = append(outFields, "#"+field)
		}
	}
	return strings.Join(outFields, ", ")
}
