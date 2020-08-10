//
package white_test

import (
	. "github.com/andrew-torda/seq_compat/pkg/white"
	"testing"
)

// TestWhiteRemove
func TestWhiteRemove(t *testing.T) {
	want := "abcdefghijk"
	ss := []string{
		want,
		" a b c d e f g h i j k",
		"a b c de fgh ijk",
		"   abcdefghijk    ",
		"a   b      cdefghijk\n ",
		"a  b  c  d   e    f     ghijk",
		"a bcdefghij   k",
		"abcdefghij\nk",
	}
	for i, s := range ss {
		b := []byte(s)
		Remove(&b)
		if string(b) != want {
			t.Fatalf("white remove broke on \"%s\" got \"%s\"", string(s), ss[i])
		}
	}
}
