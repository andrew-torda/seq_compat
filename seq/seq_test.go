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
		t.Fatal("tempfile", f_tmp, err)
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
		Vbsty: 0, Keep_gaps_rd: false,
		Dry_run:      true,
		Rmv_gaps_wrt: true}
	var names = []string{f_tmp.Name()}
	seqgrp, n_dup, err := Readfiles(names, s_opts)
	if err != nil {
		t.Fatal("Reading seqs failed", err)
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
			t.Fatalf("Seq length expected %d, got %d", seq_lengths[i], s.Testsize())
		}
	}

}

// Check that broken files are gracefully handled
func TestEmpty(t *testing.T) {
	bad_contents := []string{
		"> blah\n",
		"",
		"rubbish",
	}
	for _, content := range bad_contents {
		f_tmp, err := ioutil.TempFile("", "_del_me_testing")
		if err != nil {
			t.Fatal("tempfile", f_tmp, err)
		}
		defer os.Remove(f_tmp.Name())
		if _, err := io.WriteString(f_tmp, content); err != nil {
			t.Fatal("writing string to temp file")
		}

		f_tmp.Close()
		s_opts := &Options{Vbsty: 0, Keep_gaps_rd: true, Dry_run: true}
		if _, _, err := Readfile(f_tmp.Name(), s_opts); err == nil {
			t.Fatal("should generate error on zero-length file")
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

func TestTypes(t *testing.T) {
	var s_opts = &Options{
		Vbsty: 0, Keep_gaps_rd: false,
		Dry_run:      true,
		Rmv_gaps_wrt: true,
	}

	for tnum, x := range stypedata {
		f_tmp, err := ioutil.TempFile("", "_del_me_testing")
		if err != nil {
			t.Fatal("tempfile", f_tmp, err)
		}
		defer os.Remove(f_tmp.Name())

		if _, err := io.WriteString(f_tmp, x.s1); err != nil {
			t.Fatal("writing string to temp file")
		}
		f_tmp.Close()
		seqgrp, _, err := Readfile(f_tmp.Name(), s_opts)
		seqgrp.Upper()
		st := seqgrp.GetType()
		if st != x.stype {
			t.Fatalf("seq num %d (from 0) got %d expected %d", tnum, st, x.stype)
		}
	}
}

// roughEql says if two numbers are roughly the same
func roughEql(a, b float32) bool {
	const eps float32 = 0.01
	d := a - b
	if d < 0 {
		d = -d
	}
	if d < eps {
		return true
	}
	return false
}

// sliceEql returns true if two slices are roughly equal
func sliceEql(a, b []float32) bool {
	for i := range a {
		if roughEql(a[i], b[i]) == false {
			return false
		}
	}
	return true
}

var entdata = []struct {
	s1         string
	gapAsChar  []float32 // If gaps are characters
	gapNotChar []float32 // and if they are not counted
}{
	{`> s1'
A
> s2
C
> s3
G
> s4
T
> s5
-`,
		[]float32{1},
		[]float32{1}},
	{`> s1
AAAA
> s2
AACT
> s3
ACG-
`,
		[]float32{0.0, 0.39548847, 0.6826062, 0.6826062},
		[]float32{0.0, 0.459147917027, 0.792481250362, 0.5},
	},
	{`> s1
AAAAA
> s2
AADAC
> s3
AEG--`,
		[]float32{0, 0.20906864, 0.3608488, 0.20906864, 0.3608488},
		[]float32{0, 0.21247365, 0.3667258, 0, 0.23137821316},
	},
}

func wrtTmp(s string) (string, error) {
	f_tmp, err := ioutil.TempFile("", "_del_me_testing")
	if err != nil {
		return "", fmt.Errorf("tempfile fail")
	}

	if _, err := io.WriteString(f_tmp, s); err != nil {
		return "", fmt.Errorf("writing string to temp file %v", f_tmp.Name())
	}
	name := f_tmp.Name()
	f_tmp.Close()
	return name, nil
}

func TestEntropy(t *testing.T) {
	s_opts := &Options{
		Vbsty: 0, Keep_gaps_rd: true,
		Dry_run:      true,
		Rmv_gaps_wrt: false}

	for tnum, x := range entdata {
		var tmpname string
		var err error
		var entrpy []float32
		var seqgrp SeqGrp
		if tmpname, err = wrtTmp(x.s1); err != nil {
			t.Fatal("tempfile error:", err)
		}
		defer os.Remove(tmpname)
		if seqgrp, _, err = Readfile(tmpname, s_opts); err != nil {
			t.Fatal("Test: ", tnum, err)
		}
		seqgrp.Upper()
		if entrpy, err = seqgrp.Entropy(true); err != nil {
			t.Fatal("entropy fail on sequence", tnum)
		}

		if !sliceEql(entrpy, x.gapAsChar) {
			t.Fatal("set ", tnum, "gapaschar wanted\n", x.gapAsChar, "got\n", entrpy)
		} // now do the negative case

		seqgrp.Clear()
		if entrpy, err = seqgrp.Entropy(false); err != nil {
			t.Fatal("entropy fail on sequence", tnum)
		}

		if !sliceEql(entrpy, x.gapNotChar) {
			t.Fatal("set ", tnum, "gapnochar wanted\n", x.gapNotChar, "got\n", entrpy)
		}
	}
}

// TestFindNdx
func TestFindNdx(t *testing.T) {
	set1 := `>should be in 0
ABC
>   some stuff here for seq1
DEF
> more here in seq2
DEF`
	s_opts := &Options{
		Vbsty: 0, Keep_gaps_rd: true,
		Dry_run:      true,
		Rmv_gaps_wrt: false}
	var tmpname string
	var err error
	var seqgrp SeqGrp
	if tmpname, err = wrtTmp(set1); err != nil {
		t.Fatal("tempfile error:", err)
	}
	defer os.Remove(tmpname)
	if seqgrp, _, err = Readfile(tmpname, s_opts); err != nil {
		t.Fatalf("Reading temp sequences %v", err)
	}
	subs := []string{"> some", " some", "some", "seq1"}
	for _, s := range subs {
		if n := seqgrp.FindNdx(s); n != 1 {
			t.Fatalf("substring fail looking for %s, expected 1, got %d", s, n)
		}
	}
	if n := seqgrp.FindNdx("this string is nowhere"); n != -1 {
		t.Fatal("Failed with this string is nowhere")
	}
	if n := seqgrp.FindNdx("in seq2"); n != 2 {
		t.Fatal("Failed lookin in seq2")
	}
}

	// 1, 1/2, 0
var ufset1 string = `> reference sequence
ABC
>   some stuff here for seq1
ABD
> more here in seq2
AFE`

	// now 0, 0, 1, 1, 1, 1, 0
var ufset2 = `> reference sequence
X-BAA A-JKL
> s2
-DBAA --QKL
> s3
-GBA- --QQL
> s4
-JB-- --QQM`

func TestCompat(t *testing.T) {
	var expected = []struct {
		s string
		v []float32
	}{
		{ufset2, []float32{0, 0, 1, 1, 1, 0, 0, 0, 0.333, 0.667}},
		{ufset1, []float32{1, 0.5, 0}},
	}
	s_opts := &Options{
		Vbsty: 0, Keep_gaps_rd: true,
		Dry_run:      true,
		Rmv_gaps_wrt: false}
	var tmpname string
	var err error
	var seqgrp SeqGrp
	for i, exp := range expected {
		if tmpname, err = wrtTmp(exp.s); err != nil {
			t.Fatal("tempfile error:", err)
		}
		defer os.Remove(tmpname)
		if seqgrp, _, err = Readfile(tmpname, s_opts); err != nil {
			t.Fatalf("Reading temp sequences %v", err)
		}
		var n int
		if n = seqgrp.FindNdx("reference"); n != 0 {
			t.Fatalf("substring fail looking for %s, expected 1, got %d", "reference", n)
		}

		slc := seqgrp.GetSeqSlc()
		sq := slc[n].GetSeq()
		compat := seqgrp.Compat(sq, false)
		if !sliceEql(compat, exp.v) {
			t.Fatal("Set", i, "expected", exp.v, "got", compat)
		}
	}
}
