// Make sure we go through the greek paper and get all of their tests out.
// Then get the same tests from the Altschul paper.
package altschul_test

import (
	"github.com/andrew-torda/goutil/altschul"
	"testing"
)

func TestAlt (t *testing.T) {
	s1 := "abcde"
	s2 := "bc"
	altschul.Altschul ([]byte(s1), []byte(s2))
}
