package prago

func (resourceData *resourceData) allowsMultipleActions(user *user) (ret bool) {
	if resourceData.app.authorize(user, resourceData.canDelete) {
		ret = true
	}
	if resourceData.app.authorize(user, resourceData.canUpdate) {
		ret = true
	}
	return ret
}

func (resourceData *resourceData) getMultipleActions(user *user) (ret []listMultipleAction) {
	if !resourceData.allowsMultipleActions(user) {
		return nil
	}

	if resourceData.app.authorize(user, resourceData.canUpdate) {
		ret = append(ret, listMultipleAction{
			ID:   "edit",
			Name: "Upravit",
		})
	}

	if resourceData.app.authorize(user, resourceData.canCreate) {
		ret = append(ret, listMultipleAction{
			ID:   "clone",
			Name: "Naklonovat",
		})
	}

	ret = append(ret, resourceData.getMultipleActionsFromQuickActions(user)...)

	if resourceData.app.authorize(user, resourceData.canDelete) {
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
