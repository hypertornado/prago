package messages

import (
	"fmt"
)

var Messages *messages

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
	loc, ok := m.m[lang]
	if !ok {
		return id
	}

	str, ok := loc[id]
	if !ok {
		return id
	}

	return fmt.Sprintf(str, params...)
}
