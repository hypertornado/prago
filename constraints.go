package prago

import (
	"regexp"
)

func constraintInt(item string) func(map[string]string) bool {
	reg, _ := regexp.Compile("^[1-9][0-9]*$")
	f := constraintRegexp(item, reg)
	return f
}

func constraintWhitelist(item string, allowedValues []string) func(map[string]string) bool {
	allowedMap := make(map[string]bool)
	for _, v := range allowedValues {
		allowedMap[v] = true
	}
	return func(m map[string]string) bool {
		if value, ok := m[item]; ok {
			return allowedMap[value]
		}
		return false
	}
}

func constraintRegexp(item string, reg *regexp.Regexp) func(map[string]string) bool {
	return func(m map[string]string) bool {
		if value, ok := m[item]; ok {
			return reg.Match([]byte(value))
		}
		return false
	}
}
