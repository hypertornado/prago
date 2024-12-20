package pragosearch

import (
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// isMn checks if a rune is a non-spacing mark
func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: non-spacing marks
}

func removeDiacritics(input string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, input)
	return result
}

func lowercaser(input string) string {
	return strings.ToLower(input)
}

func tokenizer(input string) []string {
	fields := strings.FieldsFunc(input, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
	return fields
}
