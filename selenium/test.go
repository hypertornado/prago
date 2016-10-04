package selenium

import (
	"testing"
)

type Test struct {
	test *testing.T
}

func (t *Test) err(err error) {
	if err != nil {
		t.test.Fatal(err)
	}
}

func (t *Test) getStr(str string, err error) string {
	t.err(err)
	return str
}

func (t *Test) getBool(b bool, err error) bool {
	t.err(err)
	return b
}

func (t *Test) getIface(i interface{}, err error) interface{} {
	t.err(err)
	return i
}

func (t *Test) getEl(el *WebElement, err error) *WebElementTest {
	t.err(err)
	return &WebElementTest{t, el}
}

func (t *Test) getEls(els []*WebElement, err error) []*WebElementTest {
	t.err(err)
	ret := []*WebElementTest{}
	for _, v := range els {
		ret = append(ret, &WebElementTest{t, v})
	}
	return ret
}

func (t *Test) getInt(i int, err error) int {
	t.err(err)
	return i
}

func (t *Test) getInts(i, j int, err error) (int, int) {
	t.err(err)
	return i, j
}
