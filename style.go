package prago

const (
	StyleAccented = "accented"
	StyleCreate   = "create"
	StyleDestroy  = "destroy"

	blackColor = "444444"
	redColor   = "cb2431"
	greenColor = "006400"
)

func getStyleColor(style string) string {
	if style == StyleAccented {
		return "base"
	}

	if style == StyleCreate {
		return greenColor
	}

	if style == StyleDestroy {
		return redColor
	}

	return blackColor

}
