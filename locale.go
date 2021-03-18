package prago

import (
	"golang.org/x/text/language"
)

var (
	supportedLocales = []language.Tag{language.Czech, language.English}
	languageMatcher  = language.NewMatcher(supportedLocales)
	localeNames      = map[string]string{
		"cs": "Čeština",
		"en": "English",
	}
	availableLocales = [][2]string{{"en", "English"}, {"cs", "Čeština"}}
)

func validLocale(in string) bool {
	_, ok := localeNames[in]
	return ok
}

func localeFromRequest(request *Request) string {
	return localeFromAcceptLanguageString(
		request.Request().Header.Get("Accept-Language"),
	)
}

func localeFromAcceptLanguageString(acceptHeader string) string {
	t, _, _ := language.ParseAcceptLanguage(acceptHeader)
	tag, _, _ := languageMatcher.Match(t...)
	base, _ := tag.Base()

	_, ok := localeNames[base.String()]
	if ok {
		return base.String()
	}
	return availableLocales[0][0]
}

//Unlocalized creates non localized name
func Unlocalized(name string) func(string) string {
	return func(string) string {
		return name
	}
}
