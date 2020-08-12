//

package white_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/andrew-torda/seq_compat/pkg/white"
)

type fw func(*[]byte)

func benchmarkWhite(f fw, b *testing.B) {
	var tt []byte
	s := strings.Repeat("abcdefghij ", 6) + "\n"
	s = strings.Repeat(s, 10)

	for i := 0; i < 10000; i++ {
		tt = []byte(s)

		f(&tt)
		fmt.Fprintf(ioutil.Discard, "%s", tt)
	}
}

func BenchmarkByBlock(b *testing.B) {
	f := white.RemoveByBlock
	benchmarkWhite(f, b)
}
func BenchmarkByByte(b *testing.B) {
	f := white.Remove
	benchmarkWhite(f, b)
}
func BenchmarkByFields(b *testing.B) {
	f := white.RemoveByFields
	benchmarkWhite(f, b)
}
