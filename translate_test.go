package prago

import (
	"testing"
)

func TestI18N(t *testing.T) {
	translator := NewI18N()

	translator.AddLanguage("en", "i18n_test_data_en.json")
	translator.AddLanguage("", "i18n_test_data_default.json")

	data := [][]string{
		[]string{"en", "name", "name_en"},
		[]string{"", "name", "name_def"},
		[]string{"en", "nottranslated", "something_def"},
		[]string{"en", "nonsence", ""},
	}

	for _, v := range data {
		if translator.GetTranslation(v[0], v[1]) != v[2] {
			t.Errorf("translation of %s in %s is %s instead of %s", v[1], v[0], translator.GetTranslation(v[0], v[1]), v[2])
		}
	}
}
