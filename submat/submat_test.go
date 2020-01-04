package submat_test

import (
	"fmt"
	"github.com/andrew-torda/goutil/submat"
	"testing"
)

func TestA (t *testing.T) {
	const fname = "blosum62.txt"

	smat, err := submat.Read(fname)
	if err != nil {
		t.Fatal(err)
	}
	b := []byte{'a', 'C', 'w'}
	s := ""
	for _, x := range b {
		for _, y := range b {
			s += fmt.Sprint(string(x), " ", string(y), " ", smat.Score(x, y))
		}
	}
	if s != "a a 4a C 0a w -3C a 0C C 9C w -2w a -3w C -2w w 11" {
		t.Fatal ("Got wrong score string from matrix", s)
	}
}
