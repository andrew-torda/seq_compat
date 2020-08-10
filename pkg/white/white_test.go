//
package white_test

import (
	"testing"
	. "github.com/andrew-torda/seq_compat/pkg/white"
)

// TestWhiteRemove
func TestWhiteRemove (t *testing.T) {
	ss := []string {
		"abc",
		" a b c ",
		"  \n a\nb\nc\n",
	}
	for _, s := range ss {
		b := ByteSlice(s)
		b.WhiteRemove()
	}
}
