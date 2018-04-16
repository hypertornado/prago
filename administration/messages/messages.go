package messages

import (
	"fmt"
)

var (
	//Messages is singleton for admin translated messages
	Messages         *messages
	fallbackLanguage = "en"
)

func init() {
	m := map[string]map[string]string{
		"cs": cs,
		"en": en,
	}
	Messages = &messages{m}
}

type messages struct {
	m map[string]map[string]string
}

func (m *messages) Get(lang, id string, params ...interface{}) string {
	ret := m.GetNullable(lang, id, params...)
	if ret == nil {
		ret = m.GetNullable(fallbackLanguage, id, params...)
	}
	if ret == nil {
		return id
	}
	return *ret
}

func (m *messages) GetNullable(lang, id string, params ...interface{}) *string {
	loc, ok := m.m[lang]
	if !ok {
		return nil
	}

	str, ok := loc[id]
	if !ok {
		return nil
	}

	ret := fmt.Sprintf(str, params...)
	return &ret
}

func (m *messages) GetNameFunction(id string, params ...interface{}) func(string) string {
	return func(lang string) string {
		return m.Get(lang, id, params...)
	}
}