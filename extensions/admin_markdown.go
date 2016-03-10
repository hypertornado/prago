package extensions

import (
	"github.com/golang-commonmark/markdown"
)

func Markdown(in string) string {
	md := markdown.New()
	return md.RenderToString([]byte(in))
}
