package prago

import (
	"testing"
)

func TestI18N(t *testing.T) {
	test := NewTest(t)
	translator := NewI18N()

	translator.AddLanguage("en", "i18n_test_data_en.json")
	translator.AddLanguage("", "i18n_test_data_default.json")

	test.EqualString(translator.GetTranslation("en", "name"), "name_en")
	test.EqualString(translator.GetTranslation("", "name"), "name_def")
	test.EqualString(translator.GetTranslation("en", "nottranslated"), "something_def")
	test.EqualString(translator.GetTranslation("en", "nonsence"), "")
}
