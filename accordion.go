package prago

import "html/template"

type Accordion struct {
	Icon string
	Name string
	Text template.HTML
}
