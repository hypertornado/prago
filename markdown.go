package prago

import (
	"bytes"
	"html/template"
	"strings"

	stripmd "github.com/hypertornado/go-strip-markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var goldmarkConverter = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.DefinitionList,
		extension.Footnote,
		extension.Typographer,
		extension.CJK,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
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

func markdownToHTML(in string) (template.HTML, error) {
	var buf bytes.Buffer
	if err := goldmarkConverter.Convert([]byte(in), &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
