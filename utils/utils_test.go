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
		{"žšč", 3, "žšč"},
		{"žšč řďť ňěóireowprieow", 6, "žšč…"},
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
	for k, v := range [][2]string{
		{"ab**c**d", "abcd"},
		{"a\nb", "a b"},
		{"a\n\nb", "a  b"},
		{"a[b](/xx)c", "abc"},
		{"ka\n\n# b\nc", "ka  b c"},
		//{"a --- b", "a — b"},
		{"## popis aktuality", "popis aktuality"},
	} {
		if filterMarkdown(v[0]) != v[1] {
			t.Fatal(fmt.Printf("%d: Expected %s, got %s\n", k, v[1], filterMarkdown(v[0])))
		}
	}

}

func TestHumanizeFloat(t *testing.T) {
	type floatTest struct {
		f        float64
		locale   string
		expected string
	}

	for k, v := range []floatTest{
		{1.4, "cs", "1,4"},
		{1.4, "en", "1.4"},
		{1.45678, "cs", "1,45678"},
		{100000.45678, "cs", "100 000,45678"},
	} {
		res := HumanizeFloat(v.f, v.locale)
		if res != v.expected {
			t.Fatal(fmt.Printf("%d: Expected %s, got %s\n", k, v.expected, res))
		}
	}

}
