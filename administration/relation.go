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

func (r *Resource) AddRelation(r2 *Resource, field string, addName func(string) string) {
	field = columnName(field)
	r.relations = append(r.relations, relation{r2, field, addName})
}

func (resource *Resource) bindRelationActions(r relation) {
	action := Action{
		Name: r.resource.Name,
		URL:  r.resource.ID,
		Auth: r.resource.Authenticate,
		Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
			listData, err := r.resource.getListHeader(admin, user)
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
			must(admin.Query().WhereIs("id", int64(id)).Get(item))

			navigation := admin.getItemNavigation(resource, user, item, r.resource.ID)
			navigation.Wide = true

			renderNavigationPage(request, AdminNavigationPage{
				Navigation:   navigation,
				PageTemplate: "admin_list",
				PageData:     listData,
			})
		},
	}
	resource.AddItemAction(action)

	addAction := Action{
		Name: r.addName,
		URL:  "add-" + r.resource.ID,
		Auth: r.resource.Authenticate,
		Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
			values := make(url.Values)
			values.Set(r.field, request.Params().Get("id"))
			request.Redirect(resource.GetURL("new"))
		},
	}
	resource.AddItemAction(addAction)
}
