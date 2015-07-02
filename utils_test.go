package prago

import (
	"testing"
)

func TestUtils(t *testing.T) {
	test := NewTest(t)

	test.EqualString(PrettyUrl("hello"), "hello")
	test.EqualString(PrettyUrl("  Šíleně žluťoučký kůň úpěl   ďábelské ódy.  "), "silene-zlutoucky-kun-upel-dabelske-ody")

}
