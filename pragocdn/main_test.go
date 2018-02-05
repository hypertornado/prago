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
