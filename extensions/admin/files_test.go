package admin

import (
	"testing"
)

func TestFiles(t *testing.T) {
	f := &File{}
	f.Name = "ABC.jpg"
	f.UID = "abcdefgh"
	folder, file := f.GetPath("x")

	if folder != "x/a/b/c/d/e" {
		t.Fatal(folder)
	}

	if file != "x/a/b/c/d/e/fgh-ABC.jpg" {
		t.Fatal(file)
	}

}
