package prago

func RenderAction(status int, template string) func(Request) {

	return func(r Request) {
		Render(r, status, template)
	}

}
