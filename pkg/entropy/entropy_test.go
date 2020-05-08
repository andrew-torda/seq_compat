// 27 April 2020

package entropy_test

import (
	"github.com/andrew-torda/goutil/seq/common"
	. "github.com/andrew-torda/seq_compat/pkg/entropy"
	"os"
	"testing"
)

// The real testing for the calculations is in the directory with the
// entropy calculation.

// Write some junk to a file, make sure main opens it and does the right
// thing. Return the name of the file.

var seqstring string = `>s1
ACGT
> s2
-CGT
> s3
-GGT`

// check case where there are no gaps
var seqstring2 string = `> s1
ACGT
>s2
ACGT
`


func TestMain1(t *testing.T) {
	var fname string
	var err error
	runs := []string {seqstring, seqstring2}
	for _, s := range runs {
		if fname, err = common.WrtTemp(s); err != nil {
			t.Fatal("Fail writing test file")
		}
		defer os.Remove(fname)
		flags := CmdFlag{}
		if err := Mymain(&flags, fname, ""); err != nil {
			t.Fatal("bust on simple test", err)
		}
	}

}
