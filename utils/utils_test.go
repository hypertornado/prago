package utils

import (
	"fmt"
	"testing"
)

func TestPrettyURL(t *testing.T) {

	for _, v := range [][2]string{
		{"hello", "hello"},
		{"  Šíleně žluťoučký kůň úpěl   ďábelské ódy.  ", "silene-zlutoucky-kun-upel-dabelske-ody"},
	} {
		if PrettyURL(v[0]) != v[1] {
			t.Errorf("pretty url of '%s' is '%s' instead of '%s", v[0], PrettyURL(v[0]), v[1])
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

func TestFilenames(t *testing.T) {
	for _, v := range [][2]string{
		{"abc", "abc"},
		{"žluťoučký.kůň", "zlutoucky.kun"},
	} {
		if PrettyFilename(v[0]) != v[1] {
			t.Fatal(fmt.Printf("Expected %s, got %s", v[1], PrettyFilename(v[0])))
		}
	}
}

func TestFilterMarkdown(t *testing.T) {
	for _, v := range [][2]string{
		{"ab**c**d", "abcd"},
		{"a\nb", "a b"},
		{"a\n\nb", "a b"},
		{"a[b](/xx)c", "abc"},
		{"ka\n\n# b\nc", "ka b c"},
	} {
		if filterMarkdown(v[0]) != v[1] {
			t.Fatal(fmt.Printf("Expected %s, got %s\n", v[1], filterMarkdown(v[0])))
		}
	}

}
