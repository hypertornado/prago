package utils

import (
	"fmt"
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

func TestCrop(t *testing.T) {

	for k, v := range []struct {
		in    string
		index int
		out   string
	}{
		{"žšč", 2, "žšč"},
		{"žšč řďť ňěóireowprieow", 6, "žšč"},
		{"", 6, ""},
	} {
		croped := Crop(v.in, v.index)
		if croped != v.out {
			t.Fatal(fmt.Sprintf("%d Expected '%s', got '%s'", k, v.out, croped))
		}
	}

}
