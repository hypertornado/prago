package administration

import (
	"github.com/hypertornado/prago"
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

//GetLocale from request
func getLocale(request prago.Request) string {
	user, hasUser := request.GetData("currentuser").(*User)
	if hasUser {
		if validLocale(user.Locale) {
			return user.Locale
		}
	}
	return localeFromRequest(request)
}

func validLocale(in string) bool {
	_, ok := localeNames[in]
	return ok
}

func localeFromRequest(request prago.Request) string {
	acceptHeader := request.Request().Header.Get("Accept-Language")

	t, _, _ := language.ParseAcceptLanguage(acceptHeader)
	tag, _, _ := languageMatcher.Match(t...)
	return tag.String()
}
