package prago

const (
	styleAccented = "accented"
	styleCreate   = "create"
	styleDestroy  = "destroy"

	blackColor = "444444"
	redColor   = "cb2431"
	greenColor = "006400"
)

func getStyleColor(style string) string {
	if style == styleAccented {
		return "base"
	}

	if style == styleCreate {
		return greenColor
	}

	if style == styleDestroy {
		return redColor
	}

	return blackColor

}
