package admin

import (
	"github.com/hypertornado/prago"
	"strconv"
)

type relation struct {
	resource *Resource
	field    string
	addName  func(string) string
}

func (r *Resource) AddRelation(r2 *Resource, field string, addName func(string) string) {
	r.relations = append(r.relations, relation{r2, field, addName})
}

func (resource *Resource) bindRelationActions(r relation) {
	action := Action{
		Name: r.resource.Name,
		Url:  r.resource.ID,
		Auth: r.resource.Authenticate,
		Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
			listData, err := r.resource.getListHeader(admin, user)
			if err != nil {
				if err == ErrItemNotFound {
					render404(request)
					return
				}
				panic(err)
			}

			id, err := strconv.Atoi(request.Params().Get("id"))
			prago.Must(err)

			listData.PrefilterField = r.field
			listData.PrefilterValue = request.Params().Get("id")

			var item interface{}
			resource.newItem(&item)
			prago.Must(admin.Query().WhereIs("id", int64(id)).Get(item))

			navigation := admin.getItemNavigation(resource, user, item, id, r.resource.ID)
			navigation.Wide = true

			renderNavigationPage(request, AdminNavigationPage{
				Navigation:   navigation,
				PageTemplate: "admin_list",
				PageData:     listData,
			})
		},
	}
	resource.AddItemAction(action)

	if r.addName == nil {
		return
	}

	addAction := Action{
		Name: r.addName,
		Url:  "add-" + r.resource.ID,
		Auth: r.resource.Authenticate,
		Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
			prago.Redirect(request, resource.GetURL("new"))
		},
	}
	resource.AddItemAction(addAction)
}
