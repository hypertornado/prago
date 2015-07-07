package prago

import (
	"reflect"
	"runtime/debug"
	"testing"
)

type Test struct {
	t *testing.T
}

func NewTest(t *testing.T) *Test {
	return &Test{t}
}

func (t *Test) Equal(compare, expected interface{}) {
	if !reflect.DeepEqual(compare, expected) {
		t.t.Errorf("expected '%s' to be '%s'", compare, expected)
		debug.PrintStack()
	}
}

func (t *Test) EqualString(compare, expected string) {
	if compare != expected {
		t.t.Errorf("expected '%s', was '%s'", expected, compare)
		debug.PrintStack()
	}
}

func (t *Test) EqualInt(compare, expected int) {
	if compare != expected {
		t.t.Errorf("expected '%d', was '%d'", expected, compare)
		debug.PrintStack()
	}
}

func (t *Test) EqualInt64(compare, expected int64) {
	if compare != expected {
		t.t.Errorf("expected '%d', was '%d'", expected, compare)
		debug.PrintStack()
	}
}

func (t *Test) EqualFloat64(compare, expected, tolerance float64) {
	if compare < expected-tolerance || compare > expected+tolerance {
		t.t.Errorf("expected '%f', was '%f' (with tolerance %f)", expected, compare, tolerance)
		debug.PrintStack()
	}
}

func isNil(actual interface{}) bool {
	if actual == nil {
		return true
	}

	value := reflect.ValueOf(actual)
	kind := value.Kind()
	nilable := kind == reflect.Slice ||
		kind == reflect.Chan ||
		kind == reflect.Func ||
		kind == reflect.Ptr ||
		kind == reflect.Map
	return nilable && value.IsNil()
}

func (t *Test) EqualNil(compare interface{}) {
	if !isNil(compare) {
		t.t.Errorf("expected '%s' to be nil", compare)
		debug.PrintStack()
	}
}

func (t *Test) EqualNotNil(compare interface{}) {
	if isNil(compare) {
		t.t.Errorf("expected '%s' not to be nil", compare)
		debug.PrintStack()
	}
}

func (t *Test) EqualTrue(compare bool) {
	if compare != true {
		t.t.Errorf("expected '%s' to be true", compare)
		debug.PrintStack()
	}
}

func (t *Test) EqualFalse(compare bool) {
	if compare != false {
		t.t.Errorf("expected '%s' to be false", compare)
		debug.PrintStack()
	}
}
