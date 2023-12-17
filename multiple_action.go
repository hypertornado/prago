package prago

func (resource *Resource) allowsMultipleActions(userData UserData) (ret bool) {
	if userData.Authorize(resource.canDelete) {
		ret = true
	}
	if userData.Authorize(resource.canUpdate) {
		ret = true
	}
	return ret
}

func (resource *Resource) getMultipleActions(userData UserData) (ret []listMultipleAction) {
	if !resource.allowsMultipleActions(userData) {
		return nil
	}

	if userData.Authorize(resource.canUpdate) {
		ret = append(ret, listMultipleAction{
			ID:   "edit",
			Name: "Upravit",
		})
	}

	if userData.Authorize(resource.canCreate) {
		ret = append(ret, listMultipleAction{
			ID:   "clone",
			Name: "Naklonovat",
		})
	}

	if userData.Authorize(resource.canDelete) {
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
