package prago

import (
	"fmt"
)

func (resource *Resource[T]) allowsMultipleActions(user *user) (ret bool) {
	if resource.app.authorize(user, resource.canDelete) {
		ret = true
	}
	if resource.app.authorize(user, resource.canUpdate) {
		ret = true
	}
	return ret
}

func (resource *Resource[T]) getMultipleActions(user *user) (ret []listMultipleAction) {
	if !resource.allowsMultipleActions(user) {
		return nil
	}

	if resource.app.authorize(user, resource.canUpdate) {
		ret = append(ret, listMultipleAction{
			ID:   "edit",
			Name: "Upravit",
		})
	}

	if resource.app.authorize(user, resource.canCreate) {
		ret = append(ret, listMultipleAction{
			ID:   "clone",
			Name: "Naklonovat",
		})
	}

	if resource.app.authorize(user, resource.canDelete) {
		ret = append(ret, listMultipleAction{
			ID:       "delete",
			Name:     "Smazat",
			IsDelete: true,
		})
	}
	ret = append(ret, listMultipleAction{
		ID:   "cancel",
		Name: "Storno",
	})
	return
}

func (resource *Resource[T]) getItemURL(item interface{}, suffix string) string {
	ret := resource.getURL(fmt.Sprintf("%d", getItemID(item)))
	if suffix != "" {
		ret += "/" + suffix
	}
	return ret
}

func (app *App) getResourceByID(name string) resourceIface {
	return app.resourceNameMap[columnName(name)]
}

func initResource[T any](resource *Resource[T]) {
	resource.resourceController.addAroundAction(func(request *Request, next func()) {
		if !resource.app.authorize(request.user, resource.canView) {
			render403(request)
		} else {
			next()
		}
	})
}

func (resource *Resource[T]) getURL(suffix string) string {
	url := resource.id
	if len(suffix) > 0 {
		url += "/" + suffix
	}
	return resource.app.getAdminURL(url)
}

func (resource *Resource[T]) cachedCountName() string {
	return fmt.Sprintf("prago-resource_count-%s", resource.id)
}

func (resource *Resource[T]) getCachedCount() int64 {
	return resource.app.cache.Load(resource.cachedCountName(), func() interface{} {
		count, _ := resource.Query().Count()
		return count
	}).(int64)
}

func (resource *Resource[T]) updateCachedCount() error {
	count, _ := resource.Query().Count()
	return resource.app.cache.set(resource.cachedCountName(), count)
}
