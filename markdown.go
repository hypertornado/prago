package prago

import (
	"strings"

	stripmd "github.com/hypertornado/go-strip-markdown"
)

func filterMarkdown(in string) string {
	in = stripmd.StripOptions(in, stripmd.Options{
		SkipHTMLContent: true,
	})
	in = strings.Replace(in, "\n", " ", -1)
	return in
}

// CropMarkdown remove all markdown special characters
func cropMarkdown(text string, count int) string {
	text = filterMarkdown(text)
	return crop(text, count)
}
