package prago

import (
	"testing"
)

type someType struct {
	name string
}

func TestTestingFramework(t *testing.T) {
	test := NewTest(t)

	var pointer *someType = nil
	pointer = nil

	nilValues := []interface{}{nil, pointer}
	for _, v := range nilValues {
		test.EqualNil(v)
	}

	pointer = &someType{}
	notNilValues := []interface{}{"", 0, pointer}
	for _, v := range notNilValues {
		test.EqualNotNil(v)
	}

	test.EqualFloat64(1, 1.2, 0.3)

}
