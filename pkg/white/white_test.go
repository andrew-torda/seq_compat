//
package white_test

import (
	"testing"
	. "github.com/andrew-torda/seq_compat/pkg/white"
)

// TestWhiteRemove
func TestWhiteRemove (t *testing.T) {
	ss := []string {
		"abcdefghijk",
		" a b c d e f g h i j k",
		"a b c de fgh ijk",
		"   abcdefghijk    ",
		"a   b      cdefghijk\n ",
		"a  b  c  d   e    f     ghijk",
		"a bcdefghij   k",
		"abcdefghij\nk",
	}
	for _, s := range ss {
		b := ByteSlice(s)
		b.WhiteRemove()
		if string(b) != "abcdefghijk" {
			t.Fatalf ("white remove broke on \"%s\"", b)
		}
	}
}
