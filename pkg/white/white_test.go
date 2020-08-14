//
package white_test

import (
	"testing"

	. "github.com/andrew-torda/seq_compat/pkg/white"
)

type ftest func(*[]byte)

// TestWhiteRemove
func testoneversion(f ftest, t *testing.T) {

	want := "abcdefghijk"
	ss := []string{
		want,
		"abcdefghijk ",
		"abcdefghijk  ",
		"abcdefghij k ",
		"abcdefghij  k ",
		"abcdefghij  k  ",
		" abcdefghij  k  ",
		"  abcdefghij  k  ",
		" a bcdefghij  k  ",
		" a b c d e f g h i j k",
		"a b c de fgh ijk",
		"   abcdefghijk    ",
		"a   b      cdefghijk\n ",
		"a  b  c  d   e    f     ghijk",
		"a bcdefghij   k",
		"abcdefghij\tk",
	}
	for i, s := range ss {
		b := []byte(s)
		f(&b)
		if string(b) != want {
			t.Fatalf("white remove broke on \"%s\" got \"%s\"", string(b), ss[i])
		}
	}
}

func TestWhiteRemove(t *testing.T) {
	fws := []ftest{RemoveWithBlocks, Remove}
	for _, f := range fws {
		testoneversion(f, t)
	}
}
