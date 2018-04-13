package utils

import (
	"regexp"
)

//ConstraintInt limits request item on numeric types
func ConstraintInt(item string) func(map[string]string) bool {
	reg, _ := regexp.Compile("^[1-9][0-9]*$")
	f := ConstraintRegexp(item, reg)
	return f
}

//ConstraintWhitelist limits request item on allowed values
func ConstraintWhitelist(item string, allowedValues []string) func(map[string]string) bool {
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

//ConstraintMap limits request item on allowed values
func ConstraintMap(item string, allowedValues map[string]bool) func(map[string]string) bool {
	return func(m map[string]string) bool {
		return allowedValues[m[item]]
	}
}

//ConstraintRegexp limits request item by regexp
func ConstraintRegexp(item string, reg *regexp.Regexp) func(map[string]string) bool {
	return func(m map[string]string) bool {
		if value, ok := m[item]; ok {
			return reg.Match([]byte(value))
		}
		return false
	}
}
