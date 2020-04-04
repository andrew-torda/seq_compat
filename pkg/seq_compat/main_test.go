package seq_compat_test

import (
	"os"
	"testing"

	. "github.com/andrew-torda/seq_compat/pkg/seq_compat"
)

func TestMain1(t *testing.T) {
	os.Args = append(os.Args, "../seqcalc/testdata/set1.fa")
	os.Args = append(os.Args, "out.b")
	if exitCode := Mymain(); exitCode != ExitSuccess {
		t.Errorf("testmain1 got %d as return", exitCode)
	}
}
