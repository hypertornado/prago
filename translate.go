package prago

import (
	"encoding/json"
	"io/ioutil"
)

type I18N struct {
	data map[string]map[string]string
}

func NewI18N() *I18N {
	ret := &I18N{make(map[string]map[string]string)}
	return ret
}

func (t *I18N) GetTranslation(languageCode, name string) string {
	langData := t.data[languageCode]
	defaultLangData := t.data[""]

	for _, item := range []map[string]string{langData, defaultLangData} {
		if item != nil {
			res := item[name]
			if res != "" {
				return res
			}
		}
	}
	return ""
}

func (t *I18N) AddLanguage(languageCode, filePath string) error {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	res := make(map[string]string)

	err = json.Unmarshal(fileData, &res)
	if err != nil {
		return err
	}

	t.data[languageCode] = res

	return nil
}
