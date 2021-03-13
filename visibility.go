package prago

type fieldFilter func(Resource, User, Field) bool

func defaultVisibilityFilter(resource Resource, user User, f Field) bool {
	permission := f.Tags["prago-view"]
	if permission != "" {
		return resource.app.Authorize(user, Permission(permission))
	}

	visible := true
	/*if f.Name == "ID" {
		visible = false
	}*/

	if f.Tags["prago-type"] == "order" {
		visible = false
	}

	visibleTag := f.Tags["prago-visible"]
	if visibleTag == "true" {
		visible = true
	}
	if visibleTag == "false" {
		visible = false
	}
	return visible
}

func defaultEditabilityFilter(resource Resource, user User, f Field) bool {
	if !defaultVisibilityFilter(resource, user, f) {
		return false
	}

	permission := f.Tags["prago-edit"]
	if permission != "" {
		return resource.app.Authorize(user, Permission(permission))
	}

	editable := true
	if f.Name == "ID" {
		editable = false
	}
	if f.Name == "CreatedAt" || f.Name == "UpdatedAt" {
		editable = false
	}

	editableTag := f.Tags["prago-editable"]
	if editableTag == "true" {
		editable = true
	}
	if editableTag == "false" {
		editable = false
	}
	return editable
}
