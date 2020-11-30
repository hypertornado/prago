package utils

import (
	"strings"

	stripmd "github.com/writeas/go-strip-markdown"
)

func filterMarkdown(in string) string {
	in = stripmd.Strip(in)
	in = strings.Replace(in, "\n", " ", -1)
	return in
}

//CropMarkdown remove all markdown special characters
func CropMarkdown(text string, count int) string {
	text = filterMarkdown(text)
	return Crop(text, count)
}
