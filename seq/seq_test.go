package seq_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/andrew-torda/goutil/seq"
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
	defer f_tmp.Close()
	switch spaces {
	case no_spaces:
		writeTest_nospaces(f_tmp)
	case with_spaces:
		writeTest_with_spaces(f_tmp)
	}
	defer os.Remove(f_tmp.Name())
	s_opts := &Options{
		Vbsty: 0, Keep_gaps: false,
		Dry_run:  true,
		Rmv_gaps: true}
	var names = []string{f_tmp.Name()}
	seqgrp, n_dup, err := Readfiles(names, s_opts)
	if err != nil {
		t.Fatalf("Reading seqs failed %v", err)
	}
	if seqgrp.GetNSeq() != len(seq_lengths) {
		t.Fatalf("Wrote %d seqs, but read only %d.\n%s, %d",
			len(seq_lengths), seqgrp.GetNSeq(),
			"Spaces was set to ", spaces)
	}
	if n_dup != 0 {
		t.Fatalf("Found %d dups. Expected zero", n_dup)
	}
	for i, s := range seqgrp.GetSeqSlc() {
		if s.Testsize() != seq_lengths[i] {
			t.Fatalf("Sequence length expected %d, got %d", seq_lengths[i], s.Testsize())
		}
	}

}

func TestReadSeqs(t *testing.T) {
	innerWriteReadSeqs(t, no_spaces)
	innerWriteReadSeqs(t, with_spaces)
}

type testStype struct {
	s     string
	stype SeqType
}

var stypedata = []struct {
	s1    string
	stype SeqType
}{
	{"> seq1\nac gt  \n> seq 2\nACGT-ACGT", DNA},
	{"> seq1\naaa\n>seq 2\nACGT-ACG\nT", DNA},
	{"> s1\n a c    \ng-U\n>s2\naaaa", RNA},
	{"> s\nacgu\n>ss\nacgu\n\n", RNA},
	{"> s\nACGU\n>ss\nACGT\n\n", Ntide},
	{"> s\nacgu\n>ss\nACGT\n\n", Ntide},
	{"> s1\nef", Protein},
	{"> s1\nEF", Protein},
	{"> s1\nB", Unknown},
	{"> s1\njb\n>s2\nO", Unknown},
}

func breaker() {}
func TestTypes(t *testing.T) {
	s_opts := &Options{
		Vbsty: 0, Keep_gaps: false,
		Dry_run:  true,
		Rmv_gaps: true,
	}
	for tnum, x := range stypedata {
		f_tmp, err := ioutil.TempFile("", "_del_me_testing")
		if err != nil {
			t.Fatalf("tempfile %v: %s", f_tmp, err)
		}
		defer os.Remove (f_tmp.Name())
		
		if _, err := io.WriteString(f_tmp, x.s1); err != nil {
			t.Fatalf("writing string to temp file")
		}
		f_tmp.Close()
		seqgrp, _, err := Readfile(f_tmp.Name(), s_opts)
		seqgrp.Upper()
		st := seqgrp.GetType()
		if st != x.stype {
			t.Fatalf("seq num %d (from 0) got %d expected %d", tnum, st, x.stype) }

		
	}
	
}
