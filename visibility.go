package prago

type fieldFilter func(Resource, *User, field) bool

func defaultVisibilityFilter(resource Resource, user *User, f field) bool {
	permission := f.Tags["prago-can-view"]
	if permission != "" {
		return resource.app.authorize(user, Permission(permission))
	}

	visible := true
	/*if f.Name == "ID" {
		visible = false
	}*/

	if f.Tags["prago-type"] == "order" {
		visible = false
	}
	return visible
}

func defaultEditabilityFilter(resource Resource, user *User, f field) bool {
	if !defaultVisibilityFilter(resource, user, f) {
		return false
	}

	permission := f.Tags["prago-can-edit"]
	if permission != "" {
		return resource.app.authorize(user, Permission(permission))
	}

	editable := true
	if f.Name == "ID" {
		editable = false
	}
	if f.Name == "CreatedAt" || f.Name == "UpdatedAt" {
		editable = false
	}
	return editable
}
