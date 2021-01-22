package administration

type NotificationViews struct {
	Views []NotificationView
}

type NotificationView struct {
	UUID string
	Name string
}

func notificationToNotificationView(n Notification) NotificationView {
	ret := NotificationView{
		UUID: n.UUID,
		Name: n.Name,
	}
	return ret
}

func (admin *Administration) getNotificationViews(user User) (*NotificationViews, error) {
	var notifications []*Notification
	err := admin.Query().WhereIs("IsDismissed", false).WhereIs("User", user.ID).OrderDesc("ID").Get(&notifications)
	if err != nil {
		return nil, err
	}

	ret := &NotificationViews{}

	for _, v := range notifications {
		ret.Views = append(ret.Views, notificationToNotificationView(*v))
	}

	return ret, nil
}
