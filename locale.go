package prago

import (
	"fmt"

	"golang.org/x/text/language"
)

var (
	supportedLocales = []language.Tag{language.Czech, language.English}
	languageMatcher  = language.NewMatcher(supportedLocales)
	localeNames      = map[string]string{
		"cs": "Čeština",
		"en": "English",
	}
	availableLocales = [][2]string{{"cs", "Čeština"}, {"en", "English"}}
)

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

func unlocalized(name string) func(string) string {
	return func(string) string {
		return name
	}
}

func (resource *Resource) newItemName(locale string) string {
	return fmt.Sprintf("%s „%s“", messages.GetNameFunction("admin_new")(locale), resource.singularName(locale))
}
