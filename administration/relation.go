package administration

import (
	"github.com/hypertornado/prago"
	"net/url"
	"strconv"
)

type relation struct {
	resource *Resource
	field    string
	addName  func(string) string
}

func (resource *Resource) AddRelation(relatedResource *Resource, field string, addName func(string) string) {
	resource.relations = append(resource.relations, relation{relatedResource, field, addName})
}

func (resource *Resource) bindRelationActions(r relation) {
	action := Action{
		Name:       r.resource.HumanName,
		URL:        r.resource.ID,
		Permission: r.resource.CanView,
		Handler: func(resource Resource, request prago.Request, user User) {
			listData, err := r.resource.getListHeader(user)
			if err != nil {
				if err == ErrItemNotFound {
					render404(request)
					return
				}
				panic(err)
			}

			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			listData.PrefilterField = r.field
			listData.PrefilterValue = request.Params().Get("id")

			var item interface{}
			resource.newItem(&item)
			must(resource.Admin.Query().WhereIs("id", int64(id)).Get(item))

			navigation := resource.Admin.getItemNavigation(resource, user, item, r.resource.ID)
			navigation.Wide = true

			renderNavigationPage(request, adminNavigationPage{
				Navigation:   navigation,
				PageTemplate: "admin_list",
				PageData:     listData,
			})
		},
	}
	resource.AddItemAction(action)

	addAction := Action{
		Name:       r.addName,
		URL:        "add-" + r.resource.ID,
		Permission: r.resource.CanView,
		Handler: func(resource Resource, request prago.Request, user User) {
			values := make(url.Values)
			values.Set(r.field, request.Params().Get("id"))
			request.Redirect(r.resource.GetURL("new") + "?" + values.Encode())
		},
	}
	resource.AddItemAction(addAction)
}
