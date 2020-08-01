// 31 July 2020

package randseq_test
import (
	"testing"
	"strings"
	"github.com/andrew-torda/seq_compat/pkg/randseq"
)

func TestSimple (t *testing.T) {
	var sb strings.Builder
	args := randseq.RandSeqArgs {
		Wrtr: &sb,
		Cmmt: "testing seq",
		Nseq: 5000,
		Len: 1600,
	}
	if err := randseq.RandSeqMain(&args); err != nil {
		t.Fatal(err)
	}
    if n :=strings.Count (sb.String(), ">"); n != args.Nseq {
		t.Fatal("count >, got ", n, "expected", args.Nseq)
	}
}
