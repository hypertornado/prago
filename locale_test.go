package prago

import (
	"testing"
)

//TODO: better parser for language variants
func TestLocaleParser(t *testing.T) {

	for k, v := range [][2]string{
		{"cs-u-rg-czzzzz", "cs"},
		{"", "cs"},
		{"cs", "cs"},
		{"ru", "cs"},
	} {
		result := localeFromAcceptLanguageString(v[0])
		if result != v[1] {
			t.Fatal(k, v[0])
		}
	}
}
