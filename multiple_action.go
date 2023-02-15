package prago

func (resourceData *resourceData) allowsMultipleActions(userData UserData) (ret bool) {
	if userData.Authorize(resourceData.canDelete) {
		ret = true
	}
	if userData.Authorize(resourceData.canUpdate) {
		ret = true
	}
	return ret
}

func (resourceData *resourceData) getMultipleActions(userData UserData) (ret []listMultipleAction) {
	if !resourceData.allowsMultipleActions(userData) {
		return nil
	}

	if userData.Authorize(resourceData.canUpdate) {
		ret = append(ret, listMultipleAction{
			ID:   "edit",
			Name: "Upravit",
		})
	}

	if userData.Authorize(resourceData.canCreate) {
		ret = append(ret, listMultipleAction{
			ID:   "clone",
			Name: "Naklonovat",
		})
	}

	ret = append(ret, resourceData.getMultipleActionsFromQuickActions(userData)...)

	if userData.Authorize(resourceData.canDelete) {
		ret = append(ret, listMultipleAction{
			ID:       "delete",
			Name:     "Smazat",
			IsDelete: true,
		})
	}

	//resourceData.qui

	ret = append(ret, listMultipleAction{
		ID:   "cancel",
		Name: "Storno",
	})
	return
}
