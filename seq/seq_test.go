package seq_test

import (
	. "andrew/seq"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

const (
	big       = 64 * 1024
	bigminus1 = big - 1
	bigplus1  = big + 1
)

var seq_lengths = []int{10, 30, bigminus1, big, bigplus1}

// Put funny characters into the comment lines
var trickyComments = []string{
	">a☺b☻c☹d",
	">>>",
	">",
	">a comment can end in an umlautÜ",
}

const (
	no_spaces = iota
	with_spaces
)

func writeTest_with_spaces(f_tmp *os.File) {
	const b byte = 'B'
	for i, l := range seq_lengths {
		ndx := i % len(trickyComments)
		s := trickyComments[ndx]
		fmt.Fprintln(f_tmp, s)
		for j := 0; j < l; j++ {
			switch {
			case j%11 == 1:
				fmt.Fprint(f_tmp, " ")
			case j%73 == 1:
				fmt.Fprint(f_tmp, "\n")
			case j%71 == 1:
				fmt.Fprint(f_tmp, "-")
			}
			fmt.Fprint(f_tmp, string(b))
		}
		fmt.Fprint(f_tmp, "\n")
	}
}

func writeTest_nospaces(f_tmp *os.File) {
	for i := range seq_lengths {
		fmt.Fprintln(f_tmp, "> seq", i+1, ">>")
		for j := 0; j < seq_lengths[i]; j++ {
			fmt.Fprintf(f_tmp, "%c", 'A')
		}
		fmt.Fprintf(f_tmp, "\n")
	}
}

func innerWriteReadSeqs(t *testing.T, spaces int) {
	f_tmp, err := ioutil.TempFile("", "_del_me_testing")
	if err != nil {
		t.Fatalf("tempfile %v: %s", f_tmp, err)
	}
	switch spaces {
	case no_spaces:
		writeTest_nospaces(f_tmp)
	case with_spaces:
		writeTest_with_spaces(f_tmp)
	}
	defer os.Remove(f_tmp.Name())
	s_opts := &Options{
		Vbsty: 0, Keep_gaps: false,
		Dsbl_merge: false,
		Dry_run:    true,
		Min_ovlp:   50, Rmv_gaps: true}
	var names = []string{f_tmp.Name()}
	seq_set, n_dup, err := Readfiles(names, s_opts)
	if err != nil {
		t.Fatalf("Reading seqs failed %v", err)
	}
	if len(seq_set) != len(seq_lengths) {
		t.Fatalf("Wrote %d seqs, but read only %d.\n%s, %d",
			len(seq_lengths), len(seq_set),
			"Spaces was set to ", spaces)
	}
	if n_dup != 0 {
		t.Fatalf("Found %d dups. Expected zero", n_dup)
	}
	for i, s := range seq_set {
		if s.Testsize() != seq_lengths[i] {
			t.Fatalf("Sequence length expected %d, got %d", seq_lengths[i], s.Testsize())
		}
	}

}

func TestReadSeqs(t *testing.T) {
	innerWriteReadSeqs(t, no_spaces)
	innerWriteReadSeqs(t, with_spaces)
}
