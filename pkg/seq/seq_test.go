// Bug: the KeepGapsRd flag is ignored. Fix me.
package seq_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"testing"

	. "github.com/andrew-torda/seq_compat/pkg/seq"
)

const (
	big       = 64 * 1024
	bigminus1 = big - 1
	bigplus1  = big + 1
)

var seq_lengths = []int{10, 30, bigminus1, big, bigplus1}

func cmmtHelp(got, want string, t *testing.T) {
	if got != want {
		t.Fatalf("checking comments wanted \"%s\" got \"%s\"", want, got)
	}
}

// TestComment is to check that comments are read exactly, correctly
func TestComment(t *testing.T) {
	c0 := "testcomment no space"
	c1 := " testcomment with space at start"
	s := "aaa\n"
	seqs := ">" + c0 + "\n" + s + ">" + c1 + "\n" + s
	sr := strings.NewReader(seqs)
	var seqgrp SeqGrp
	var s_opts Options

	if err := ReadFasta(sr, &seqgrp, &s_opts); err != nil {
		t.Fatal("bust reading simple seqs in TestComment", err)
	}
	slc := seqgrp.SeqSlc()

	cmmtHelp(slc[1].Cmmt(), c1, t)
	cmmtHelp(slc[0].Cmmt(), c0, t)
}

// TestDiffLen checks if we can read sequences of different lengths
func TestDiffLen(t *testing.T) {
	s := `>s1
a
> s2
aa
> s3
aa-a`
	var seqgrp SeqGrp
	s_opts := &Options{
		DiffLenSeq: true,
		RmvGapsRd:  true,
	}

	if err := ReadFasta(strings.NewReader(s), &seqgrp, s_opts); err != nil {
		t.Fatal("Reading seqs failed", err)
	}
	if ngot := seqgrp.NSeq(); ngot != 3 {
		t.Fatalf("Seqs of diff length got %d wanted 3 seqs", ngot)
	}
	for i := 0; i < 3; i++ {
		ss := seqgrp.SeqSlc()[i]
		l := ss.Len()
		if l != i+1 {
			t.Fatalf("seqs diff length got %d wanted %d", l, i+1)
		}
	}
}

// TestDiffLenLong has different length sequences that should be much longer
// than one buffer.
func TestDiffLenLong(t *testing.T) {
	ll := []int{10000, 20000, 50000}
	s := ">\n" + strings.Repeat("a", ll[0]) + "\n> s2\n" + strings.Repeat("c", ll[1]) +
		"\n> s3\n" + strings.Repeat("d", ll[2])
	var seqgrp SeqGrp
	s_opts := &Options{
		DiffLenSeq: true,
		RmvGapsRd:  true,
	}

	if err := ReadFasta(strings.NewReader(s), &seqgrp, s_opts); err != nil {
		t.Fatal("Reading seqs failed", err)
	}
	if ngot := seqgrp.NSeq(); ngot != 3 {
		t.Fatalf("Seqs of diff length got %d wanted 3 seqs", ngot)
	}
	for i := 0; i < len(ll); i++ {
		l := seqgrp.SeqSlc()[i].Len()
		if l != ll[i] {
			t.Fatalf("long seq wanted %d got %d", ll[i], l)
		}
	}
}

// TestFastaBug is to track down a specific bug I had
func TestFastaBug(t *testing.T) {
	const nseq = 5
	const sLen = 16
	sb := ""
	for i := 0; i < nseq; i++ {
		sb += fmt.Sprintf("> some %d comment\n", i)
		for j := 0; j < sLen; j++ {
			sb += fmt.Sprintf("%d", i)
		}
		sb += "\n"
	}

	SetFastaRdSize(200)

	s_opts := &Options{}

	var seqgrp SeqGrp
	if err := ReadFasta(strings.NewReader(sb), &seqgrp, s_opts); err != nil {
		t.Fatal("Reading seqs failed", err)
	}
	if seqgrp.NSeq() != nseq {
		t.Fatalf("Got %d wanted %d seqlen was %d\n", seqgrp.NSeq(), nseq, seqgrp.GetLen())
	}

}

// TestRangeShort
func TestRangeShort(t *testing.T) {
	s := `> s1
0123456789
> s2
abcdefghij
> s3
ABCDEFGHIJ`
	var seqgrp SeqGrp
	s_opts := &Options{
		RangeStart: 2,
		RangeEnd:   5,
	}
	if err := ReadFasta(strings.NewReader(s), &seqgrp, s_opts); err != nil {
		t.Fatal("Reading seqs failed", err)
	}
	if ngot := seqgrp.NSeq(); ngot != 3 {
		t.Fatalf("Seqs of diff length got %d wanted 3 seqs", ngot)
	}
	want := []string{"2345", "cdef", "CDEF"}
	for i := 0; i < len(want); i++ {
		got := seqgrp.SeqSlc()[i].GetSeq()
		if string(got) != want[i] {
			t.Fatal("got", string(got), "wanted", want[i])
		}
	}
}

// TestRangeBroken should catch invalid ranges
func TestRangeBroken(t *testing.T) {
	s := `> s1
0123456789
> s2
abcdefghij
> s3
ABCDEFGHIJ`
	var seqgrp SeqGrp
	s_opts := &Options{
		RangeStart: 2,
		RangeEnd:   10,
	}
	if err := ReadFasta(strings.NewReader(s), &seqgrp, s_opts); err == nil {
		t.Fatal("Should have failed with invalid range")
	}
}

const (
	no_spaces = iota
	with_spaces
)

// Put funny characters into the comment lines
var trickyComments = []string{
	">a☺b☻c☹d",
	">>>",
	">",
	">a comment can end in an umlautÜ",
}

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

func TestBrokenSeq(t *testing.T) {
	s := `> s1
abc
> s2 there is no sequence next`
	var seqgrp SeqGrp
	s_opts := &Options{}
	if err := ReadFasta(strings.NewReader(s), &seqgrp, s_opts); err == nil {
		t.Fatal("incomplete sequence did not break")
	}
	print()
}

// TestReadFastashort uses buffers of various lengths to catch end of buffer mistakes.
func TestReadFastaShort(t *testing.T) {
	set1 := ">\n" + "abcdefghij\n" +
		"> longer comment" + strings.Repeat(" x", 300) + "\n" +
		strings.Repeat("a", 10) + "\n" + "> longer comment" + strings.Repeat(" x", 3) +
		"\n" + strings.Repeat(" b ", 10) + strings.Repeat(" ", 167)
	bsize := []int{3, 4, 5, 10, 100, 512}

	for i, bs := range bsize {
		rdr := strings.NewReader(set1)
		var seqgrp SeqGrp
		s_opts := &Options{}
		SetFastaRdSize(bs)
		if err := ReadFasta(rdr, &seqgrp, s_opts); err != nil {
			t.Fatal(err)
		}
		if n := seqgrp.GetLen(); n != 10 {
			t.Fatal("seq num", i, "got", seqgrp.GetLen(), "want 10")
		}
		if n := seqgrp.NSeq(); n != 3 {
			t.Fatal("seq loop num", i, "got nseq", seqgrp.NSeq(), "want 3")
		}
	}
}

// innerWriteReadSeqs writes and then reads a sequence. It should be called
// once with spaces and once without.
func innerWriteReadSeqs(t *testing.T, spaces bool) {
	var b strings.Builder

	switch spaces {
	case false:
		writeTest_nospaces(&b)
	case true:
		writeTest_with_spaces(&b)
	}
	reader := strings.NewReader(b.String())

	s_opts := &Options{
		RmvGapsRd:  true,
		DryRun:     true,
		RmvGapsWrt: true,
		DiffLenSeq: true,
	}

	var seqgrp SeqGrp
	if err := ReadFasta(reader, &seqgrp, s_opts); err != nil {
		t.Fatal("Reading seqs failed", err)
	}

	if seqgrp.NSeq() != len(seq_lengths) {
		t.Fatalf("Wrote %d seqs, but read only %d.\n%s, %t",
			len(seq_lengths), seqgrp.NSeq(),
			"Spaces was set to ", spaces)
	}
	for i, s := range seqgrp.SeqSlc() {
		if !spaces {
			if s.Len() != seq_lengths[i] {
				t.Fatalf("Seq length expected %d, got %d", seq_lengths[i], s.Len())
			}
		}
	}
}

// TestEmpty checks that broken files are gracefully handled
func TestEmpty(t *testing.T) {
	bad_contents := []string{
		"> blah\n",
		"",
		"rubbish",
	}
	for _, content := range bad_contents {
		f_tmp, err := os.CreateTemp("", "_del_me_testing")
		if err != nil {
			t.Fatal("tempfile", f_tmp, err)
		}
		defer os.Remove(f_tmp.Name())
		if _, err := io.WriteString(f_tmp, content); err != nil {
			t.Fatal("writing string to temp file")
		}

		f_tmp.Close()
		s_opts := &Options{}
		if _, err := Readfile(f_tmp.Name(), s_opts); err == nil {
			t.Fatal("should generate error on zero-length file")
		}
	}

}

// TestErrorOnDiffSeqs should provoke the error when we expect sequences
// to be the same length, but they are not.
func TestErrorOnDiffSeqs(t *testing.T) {
	texts := []string{
		"> seq1\naaaa\n> seq 2\naaaaa",
		"> seq1\naaaaa\n> seq 2\naaaa",
	}

	var seqgrp SeqGrp
	var s_opts = &Options{DiffLenSeq: false}
	for _, txt := range texts {
		rdr := strings.NewReader(txt)
		if err := ReadFasta(rdr, &seqgrp, s_opts); err == nil {
			t.Fatal("Should provoke error on uneven sequences")
		}
	}
}

// TestReadFasta writes and then reads sequences, and does it once to check that
// we hop over white space and once to make sure we handle long lines.
func TestReadFasta(t *testing.T) {
	spaces := []bool{false, true}
	for _, tt := range spaces {
		innerWriteReadSeqs(t, tt)
		innerWriteReadSeqs(t, tt)
	}
}

type testStype struct {
	s     string
	stype SeqType
}

var stypedata = []struct {
	s1    string
	stype SeqType
}{
	{"> s\nACGU\n>ss\nACGT\n\n", Ntide},
	{"> seq1\nACGT-ACGT\n> seq 2\n acgt", DNA},
	{"> seq1\nac gt  \n> seq 2\nACGT-ACGT", DNA},
	{"> seq1\naaa\n>seq 2\nACGT-ACG\nT", DNA},
	{"> s1\n a c    \ng-U\n>s2\naaaa", RNA},
	{"> s\nacgu\n>ss\nacgu\n\n", RNA},
	{"> s\nacgu\n>ss\nACGT\n\n", Ntide},
	{"> s1\nef", Protein},
	{"> s1\nEF", Protein},
	{"> s1\nB", Unknown},
	{"> s1\njb\n>s2\nO", Unknown},
}

// TestTypes checks the code for recognising RNA/DNA/Protein/whatever types.
func TestTypes(t *testing.T) {
	var s_opts = &Options{
		RmvGapsRd: true,
		DiffLenSeq: true,
	}

	for tnum, x := range stypedata {
		var seqgrp SeqGrp
		if err := ReadFasta(strings.NewReader(x.s1), &seqgrp, s_opts); err != nil {
			t.Fatal("TestTypes broke on ReadFasta", err)
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
	f_tmp, err := os.CreateTemp("", "_del_me_testing")
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
	s_opts := &Options{}

	for tnum, x := range entdata {
		var seqgrp SeqGrp
		rdr := strings.NewReader(x.s1)
		if err := ReadFasta(rdr, &seqgrp, s_opts); err != nil {
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
	s_opts := &Options{}

	var seqgrp SeqGrp

	rdr := strings.NewReader(set1)
	if err := ReadFasta(rdr, &seqgrp, s_opts); err != nil {
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
	s_opts := &Options{}

	for i, exp := range expected {
		var err error
		var tmpname string
		if tmpname, err = wrtTmp(exp.s); err != nil {
			t.Fatal("tempfile error:", err)
		}
		defer os.Remove(tmpname)
		var seqgrp *SeqGrp
		if seqgrp, err = Readfile(tmpname, s_opts); err != nil {
			t.Fatalf("Reading temp sequences %v", err)
		}
		var n int
		if n = seqgrp.FindNdx("reference"); n != 0 {
			t.Fatalf("substring fail looking for %s, expected 1, got %d", "reference", n)
		}

		slc := seqgrp.SeqSlc()
		sq := slc[n].GetSeq()
		compat := seqgrp.Compat(sq, false)
		if !sliceEql(compat, exp.v) {
			t.Fatal("Set", i, "expected", exp.v, "got", compat)
		}
	}
}

// getSeqGrpSameLen returns a seqgroup in which all the sequences have the same
// length
func getSeqGrpSameLen() *SeqGrp {
	ss := []string{"aaaa", "abcd"}
	return Str2SeqGrp(ss, "s")
}

// TestStr2SeqGrp
func TestStr2SeqGrp(t *testing.T) {
	ss := []string{"aa", "bb", "cc"}
	seqgrp := Str2SeqGrp(ss, "s")
	if i := seqgrp.NSeq(); i != 3 {
		t.Fatalf("Wrong num seqs, want 3, got %d", i)
	}
	if i := len(seqgrp.SeqSlc()[0].GetSeq()); i != 2 {
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
		t.Fatal(`broken length of "a"`)
	}
	if a[0] != 'a' {
		t.Fatal(`did not find "a" in first place in revmap`)
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
}

// TestZeroLen checks that zero length sequences provoke an
// error unless overridden
func TestZeroLen (t *testing.T) {
	sdata := ">s1\nacd--\n>s2\n-----\n> s3\nefghi\n"
	var seqgrp SeqGrp
	var s_opts = &Options{
		RmvGapsRd: true,
	}

	if err := ReadFasta (strings.NewReader(sdata), &seqgrp, s_opts); err == nil {
		t.Fatal("TestZeroLen should provoke error on zero length seq")
	}
	s_opts.ZeroLenOK = true
	s_opts.DiffLenSeq = true
	if err := ReadFasta (strings.NewReader(sdata), &seqgrp, s_opts); err != nil {
		t.Fatal("TestZeroLen should have ignored zero length", err)
	}
}

// TestSeqInfo tests some seq manipulation functions
func TestSeqInfo(t *testing.T) {
	ss := []string{"aa", "bb", "cc"}
	const sometext = "sometext is here"
	seqgrp := Str2SeqGrp(ss, sometext)

	a0 := &(seqgrp.SeqSlc()[0])
	a1 := &(seqgrp.SeqSlc()[1])

	c := a0.Cmmt()
	if strings.Contains(c, sometext) == false {
		t.Fatal("did not find: " + sometext + " got " + c)
	}
	c = a0.String()
	if strings.Contains(c, sometext) == false {
		t.Fatal("in whole seq did not find:" + sometext)
	}
	if strings.Contains(c, "aa") == false {
		t.Fatal(`in whole seq did not find "aa"`)
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
	if bytes.Contains(seqgrp.SeqSlc()[0].GetSeq(), []byte(aaaaaaaa)) == false {
		t.Fatal("Did not change sequence aaaa properly")
	}
	if bytes.Contains(seqgrp.SeqSlc()[1].GetSeq(), []byte(bbbbbbbb)) == false {
		t.Fatal("Did not change sequence bbbbb properly")
	}
}
