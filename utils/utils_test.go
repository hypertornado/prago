package utils

import (
	"testing"
)

func TestUtils(t *testing.T) {

	data := [][]string{
		[]string{"hello", "hello"},
		[]string{"  Šíleně žluťoučký kůň úpěl   ďábelské ódy.  ", "silene-zlutoucky-kun-upel-dabelske-ody"},
	}

	for _, v := range data {
		if PrettyUrl(v[0]) != v[1] {
			t.Errorf("pretty url of '%s' is '%s' instead of '%s", v[0], PrettyUrl(v[0]), v[1])
		}
	}
}
