package prago

func (request *Request) AddFlashMessage(message string) {
	request.app.NewNotification(message).Flash(request)
}

func (n *Notification) Flash(request *Request) error {
	n.isFlash = true
	n.app.notificationCenter.add(n)
	request.setCookie(n.app.getFlashCookieID(), n.uuid)
	return nil
}
