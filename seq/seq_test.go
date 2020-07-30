package seq_test

import (
	"bytes"
	"fmt"
	. "github.com/andrew-torda/goutil/seq"
	"io"
	"io/ioutil"
	"os"
	"strings"
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

// writeTest_with_spaces provides some sequences with different patterns of
// white space and some gap characters mixed in. It sticks it in an io.Writer.
func writeTest_with_spaces(f_tmp io.Writer) {
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

// writeTest_nospaces puts some sequences into an io.Writer, but with no spaces
// so as to check if we correctly handle long lines.
func writeTest_nospaces(f_tmp io.Writer) {
	for _, i := range seq_lengths {
		fmt.Fprintln(f_tmp, "> seq", i+1, ">>")
		for j := 0; j < i; j++ {
			fmt.Fprintf(f_tmp, "%c", 'A')
		}
		fmt.Fprintf(f_tmp, "\n")
	}
}

// innerWriteReadSeqs writes and then reads a sequence. It should be called
// once with spaces and once without.
func innerWriteReadSeqs(t *testing.T, spaces int) {
	var b strings.Builder

	switch spaces {
	case no_spaces:
		writeTest_nospaces(&b)
	case with_spaces:
		writeTest_with_spaces(&b)
	}
	reader := strings.NewReader(b.String())

	s_opts := &Options{
		Vbsty: 0, Keep_gaps_rd: false,
		Dry_run:      true,
		Rmv_gaps_wrt: true}

	var seqgrp SeqGrp
	n_dup, err := ReadSeqs(reader, &seqgrp, s_opts)
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
		if s.Len() != seq_lengths[i] {
			t.Fatalf("Seq length expected %d, got %d", seq_lengths[i], s.Len())
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

// TestReadSeqs writes and then reads sequences, and does it once to check that
// we hop over white space and once to make sure we handle long lines.
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

// TestTypes checks the code for recognising RNA/DNA/Protein/whatever types.
func TestTypes(t *testing.T) {
	var s_opts = &Options{
		Vbsty: 0, Keep_gaps_rd: false,
		Dry_run:      true,
		Rmv_gaps_wrt: true,
	}

	for tnum, x := range stypedata {
		var seqgrp SeqGrp
		if _, err := ReadSeqs(strings.NewReader(x.s1), &seqgrp, s_opts); err != nil {
			t.Fatal("TestTypes broke on ReadSeqs", err)
		}
		seqgrp.Upper()
		st := seqgrp.GetType()
		if st != x.stype {
			const msg = "seq num %d (numbering from 0) got type %d expected %d"
			t.Fatalf(msg, tnum, st, x.stype)
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

// TestEntropy checks the entropy calculation.
func TestEntropy(t *testing.T) {
	s_opts := &Options{
		Vbsty: 0, Keep_gaps_rd: true,
		Dry_run:      true,
		Rmv_gaps_wrt: false}

	for tnum, x := range entdata {
		var seqgrp SeqGrp
		rdr := strings.NewReader(x.s1)
		if _, err := ReadSeqs(rdr, &seqgrp, s_opts); err != nil {
			t.Fatal("Test: ", tnum, err)
		}
		seqgrp.Upper()
		entrpy := make([]float32, seqgrp.GetLen())
		seqgrp.Entropy(true, entrpy)
		const emsg = "set %d %s wanted\n %f got %f\n"
		if !sliceEql(entrpy, x.gapAsChar) {
			t.Fatalf(emsg, tnum, "gap as char", x.gapAsChar, entrpy)
		} // now do the negative case

		seqgrp.Clear()
		seqgrp.Entropy(false, entrpy)

		if !sliceEql(entrpy, x.gapNotChar) {
			t.Fatalf(emsg, tnum, "gap nochar", x.gapAsChar, entrpy)
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

	var seqgrp SeqGrp

	rdr := strings.NewReader(set1)
	if _, err := ReadSeqs(rdr, &seqgrp, s_opts); err != nil {
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
	var seqgrp *SeqGrp
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

// getSeqGrpSameLen returns a seqgroup in which all the sequences have the same
// length
func getSeqGrpSameLen() SeqGrp {
	ss := []string{"aaaa", "abcd"}
	return Str2SeqGrp(ss, "s")
}

// TestStr2SeqGrp
func TestStr2SeqGrp(t *testing.T) {
	ss := []string{"aa", "bb", "cc"}
	seqgrp := Str2SeqGrp(ss, "s")
	if i := seqgrp.GetNSeq(); i != 3 {
		t.Fatalf("Wrong num seqs, want 3, got %d", i)
	}
	if i := len(seqgrp.GetSeqSlc()[0].GetSeq()); i != 2 {
		t.Fatalf("Wrong seq len, want 2, got %d", i)
	}
}

// TestGetCounts
func TestGetCounts(t *testing.T) {
	seqgrp := Str2SeqGrp([]string{"aa", "bb", "cc"}, "s")
	b := seqgrp.GetCounts().Mat
	const blen = 3
	const b0len = 2
	if len(b) != blen {
		t.Fatalf("len b want %d  got %d", blen, len(b))
	}
	if len(b[0]) != b0len {
		t.Fatalf("len b[0] want %d got %d", b0len, len(b[0]))
	}
}

// TestTypeKnwn
func TestTypeKnwn(t *testing.T) {
	seqgrp := Str2SeqGrp([]string{"aa", "bb", "cc"}, "s")
	if seqgrp.TypeKnwn() == true {
		t.Fatal("typeknown should not know what type we are")
	}
}

// TestGetRevmap
func TestGetRevmap(t *testing.T) {
	seqgrp := Str2SeqGrp([]string{"aa", "bb", "cc"}, "s")
	if len(seqgrp.GetRevmap()) != 0 {
		t.Fatal("Should not have a revmap yet")
	}
	seqgrp.UsageSite()
	a := seqgrp.GetRevmap()
	if len(a) != 3 {
		t.Fatal("broken length of \"a\"")
	}
	if a[0] != 'a' {
		t.Fatal("did not find \"a\" in first place in revmap")
	}
}

// TestGetSymUsed
func TestGetSymUsed(t *testing.T) {
	nothing := func(x bool) bool { return x }
	seqgrp := Str2SeqGrp([]string{"aa", "bb", "cc"}, "s")
	a := seqgrp.GetSymUsed()
	_ = nothing(a[MaxSym-1])
}

// TestGetNSeq
func TestGetNSeq(t *testing.T) {
	testdat := []struct {
		ss   []string
		nsym int
		tt   []string
		ncmb int
	}{
		{[]string{"a", "a"}, 1, []string{"a", "a"}, 1},
		{[]string{"ab", "ab"}, 2, []string{"cd", "ce"}, 5},
		{[]string{"abc", "def", "abc"}, 6, []string{"abcdefg"}, 7},
	}

	for _, a := range testdat {
		seqgrp := Str2SeqGrp(a.ss)
		nsym := seqgrp.GetNSym()
		if nsym != a.nsym {
			t.Fatalf("Wrong nsym. Wanted %d, got %d", a.nsym, nsym)
		}
	}
	for _, a := range testdat {
		seqgrp1 := Str2SeqGrp(a.ss)
		seqgrp2 := Str2SeqGrp(a.tt)
		symSync := SymSync{UChan: make(chan [MaxSym]bool)}
		go seqgrp1.SetSymUsed(&symSync)

		seqgrp2.SetSymUsed(&symSync)
		if seqgrp1.GetNSym() != seqgrp2.GetNSym() {
			t.Fatalf("nsym mismatch %d vs %d", seqgrp1.GetNSym(), seqgrp2.GetNSym())
		}
		if n := seqgrp1.GetNSym(); n != a.ncmb {
			t.Fatalf("combined nsyms, wanted %d got %d", a.ncmb, n)
		}
	}
}

// TestSeqInfo tests some seq manipulation functions
func TestSeqInfo(t *testing.T) {
	ss := []string{"aa", "bb", "cc"}
	const sometext = "sometext is here"
	seqgrp := Str2SeqGrp(ss, sometext)

	a0 := &(seqgrp.GetSeqSlc()[0])
	a1 := &(seqgrp.GetSeqSlc()[1])

	c := a0.GetCmmt()
	if strings.Contains(c, sometext) == false {
		t.Fatal("did not find: " + sometext + " got " + c)
	}
	c = a0.String()
	if strings.Contains(c, sometext) == false {
		t.Fatal("in whole seq did not find:" + sometext)
	}
	if strings.Contains(c, "aa") == false {
		t.Fatal("in whole seq did not find \"aa\"")
	}

	a0.Upper()
	g := a0.GetSeq()
	if bytes.Contains(g, []byte("AA")) == false {
		t.Fatal("did not uppercase sequence")
	}
	a0.Lower()
	g = a0.GetSeq()
	if bytes.Contains(g, []byte("aa")) == false {
		t.Fatal("did not lowercase sequence")
	}
	const aaaaaaaa = "aaaaaaaa"
	const bbbbbbbb = "bbbbbbbb"
	a0.SetSeq([]byte(aaaaaaaa))
	a1.SetSeq([]byte(bbbbbbbb))
	// Now go back to the slice and check if the values are there
	if bytes.Contains(seqgrp.GetSeqSlc()[0].GetSeq(), []byte(aaaaaaaa)) == false {
		t.Fatal("Did not change sequence aaaa properly")
	}
	if bytes.Contains(seqgrp.GetSeqSlc()[1].GetSeq(), []byte(bbbbbbbb)) == false {
		t.Fatal("Did not change sequence bbbbb properly")
	}
}
