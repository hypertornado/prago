package prago

type notificationViews struct {
	Views []notificationView
}

type notificationView struct {
	UUID string
	Name string
}

func notificationToNotificationView(n Notification) notificationView {
	ret := notificationView{
		UUID: n.UUID,
		Name: n.Name,
	}
	return ret
}

func (app *App) getNotificationViews(user User) (*notificationViews, error) {
	var notifications []*Notification
	err := app.Query().WhereIs("IsDismissed", false).WhereIs("User", user.ID).OrderDesc("ID").Get(&notifications)
	if err != nil {
		return nil, err
	}

	ret := &notificationViews{}

	for _, v := range notifications {
		ret.Views = append(ret.Views, notificationToNotificationView(*v))
	}

	return ret, nil
}
