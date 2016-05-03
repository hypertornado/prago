package admin

import (
	"fmt"
	"github.com/hypertornado/prago"
	"golang.org/x/text/language"
)

var (
	defaultLocale    = "cs"
	supportedLocales = []language.Tag{language.Czech, language.English}
	languageMatcher  = language.NewMatcher(supportedLocales)
	localeNames      = map[string]string{
		"cs": "Čeština",
		"en": "English",
	}
)

func Locale(request prago.Request) {
	if true {
		return
	}

	acceptHeader := request.Request().Header.Get("Accept-Language")
	println(acceptHeader)

	t, q, err := language.ParseAcceptLanguage(acceptHeader)
	tag, _, _ := languageMatcher.Match(t...)
	fmt.Println(tag.String())
	fmt.Printf("%5v (t: %6v; q: %3v; err: %v)\n", tag, t, q, err)

	//println(language.Czech.String())
	//println(language.English.String())
}
