package messages

import (
	"fmt"
	"time"
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

func (m *messages) Timestamp(lang string, t time.Time) string {
	if lang == "cs" {
		return t.Format("2. 1. 2006 15:04")
	}
	return t.Format("2006-01-02 15:04")
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
