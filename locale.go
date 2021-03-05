package prago

import (
	"golang.org/x/text/language"
)

var (
	supportedLocales = []language.Tag{language.English, language.Czech}
	languageMatcher  = language.NewMatcher(supportedLocales)
	localeNames      = map[string]string{
		"cs": "Čeština",
		"en": "English",
	}
	availableLocales = [][2]string{{"en", "English"}, {"cs", "Čeština"}}
)

//GetLocale from request
func getLocale(request Request) string {
	user, hasUser := request.GetData("currentuser").(*User)
	if hasUser {
		if validLocale(user.Locale) {
			return user.Locale
		}
	}
	return localeFromAcceptLanguageString(
		request.Request().Header.Get("Accept-Language"),
	)
}

func validLocale(in string) bool {
	_, ok := localeNames[in]
	return ok
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
