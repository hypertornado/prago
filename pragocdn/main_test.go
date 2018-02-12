package main

import (
	"testing"
)

func TestNames(t *testing.T) {

	for k, v := range [][3]string{
		{"a.b", "a", "b"},
		{"hello.world", "hello", "world"},
		{"a.b.c", "a.b", "c"},
		{"aa.jpeg", "aa", "jpg"},
		{"28dc80c83776a8d4b733d10b03d3-r16-9-w368-h206-gd468002e575111e5b3730025900fea04.jpg", "28dc80c83776a8d4b733d10b03d3-r16-9-w368-h206-gd468002e575111e5b3730025900fea04", "jpg"},
	} {
		name, ext, err := getNameAndExtension(v[0])
		if err != nil {
			t.Fatal(k, err)
		}
		if name != v[1] {
			t.Fatal(k, name, v[1])
		}
		if ext != v[2] {
			t.Fatal(k, ext, v[2])
		}
	}
}
