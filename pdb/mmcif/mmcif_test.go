package mmcif_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	. "pdb/mmcif"
	"pdb/zwrap"
	"strings"
	"testing"
)

type twostring struct {
	in  string
	out string
}

func TestMessyLine(t *testing.T) {
	// This is from 2a9w.cif. I think there should be seven pieces
	ss :=
		`GA9 non-polymer         . '3,3-BIS(3-BR-4-HYD)-7-CH-1H,3H-BEO[DE]ISO-1-ONE'
'4-CHL-3',3"-DIB-1,8-NAPHTH' 'C24 H13 Br2 Cl O4' 560.619
GLN 'L-peptide linking' y GLUTAMINE                                                                   ? 'C5 H10 N2 O3'
146.144
`
	answers := []int{4, 3, 6, 1}
	scnr := NewCmmtScanner(bytes.NewReader([]byte(ss)), '#')
	type sbyte []byte
	retIn := make([][]byte, 0, 40)
	ndx := 0
	for scnr.Cscan() == true && scnr.Cbytes() != nil {
		tt, err := SplitCifLine(scnr.Cbytes(), retIn)
		if err != nil {
			t.Error("Splitting messy string", err)
		}
		if len(tt) != answers[ndx] {
			t.Error("wrong number of entries, got", len(tt))
		}
		ndx++
	}
}

// getFpLog automates the routine steps we take in many of the tests.
// It joins the directory to the file, tries to open it, then calls a
// function in zwrap which checks if it is compressed and returns an
// appropriate ReadCloser.
func getFp(dir string, fname string, t *testing.T) (io.ReadCloser, error) {
	if dir != "" {
		fname = filepath.FromSlash(dir + "/" + fname)
	}
	fp, err := os.Open(fname)
	if err != nil {
		return nil, errors.New("opening file error")
	}
	rdr, e2 := zwrap.WrapMaybe(fp)
	if e2 != nil {
		t.Error("broke in zwrap on file ", fname, e2.Error())
		return nil, err
	}
	return rdr, nil
}

const testdata string = "testdata"

func TestMessy2(t *testing.T) {
	testfile := "lots_of_quotes.cif"
	fp, err := getFp(testdata, testfile, t)
	if err != nil {
		t.Error("unexpected error starting on " + testfile)
	}
	defer fp.Close()

	mr := NewMmcifReader(fp)
	mr.AddTable([]string{"_chem_comp."})
	md, err := mr.DoFile()
	if err != nil {
		t.Error(err)
	}
	if md == nil {
		t.Error("fail reading from", testfile)
	}
	if len(md.Tables["_chem_comp"].Names) != 7 {
		t.Error("Fail reading _chem_comp names, file", testfile)
	}
	if len(md.Tables["_chem_comp"].Vals) != 18 {
		t.Error("Fail gettings values, file", testfile)
	}
}

var cmoddata = []struct {
	chains   []string
	atoms    []string
	modelmax int16
	valid    int
	invalid  int
}{ // per model there are 12 CA atoms in chain A and 5 in Chain B
	{[]string{"A"}, []string{}, -1, 597, -1},
	{[]string{"B"}, []string{}, -1, 171, -1},
	{[]string{}, []string{}, -1, 597 + 171, -1},
	{[]string{}, []string{"CA"}, -1, 51, 0},
	{[]string{"A", "B"}, []string{"CA"}, -1, 51, 0},
	{[]string{"A", "B"}, []string{"CB"}, -1, 45, 0},
	{[]string{"A", "B"}, []string{"CA", "CB"}, -1, 96, 6},
	{[]string{"A"}, []string{"CA"}, -1, 36, 0}, // all models
	{[]string{"A"}, []string{"CA"}, 0, 0, 0},   // no models
	{[]string{"A"}, []string{"CA"}, 1, 12, 0},  // 1 model
	{[]string{"A"}, []string{"CA"}, 2, 24, 0},
	{[]string{"A"}, []string{"CA"}, 3, 36, 0},
	{[]string{"B"}, []string{"CA"}, 1, 5, 0},
}

func TestMmcifChainModels(t *testing.T) {
	testfile := "threemodel.cif"
	for _, cm := range cmoddata {
		fp, err := getFp(testdata, testfile, t)
		if err != nil {
			t.Error("Unexpected error file " + testfile)
		}
		mr := NewMmcifReader(fp)
		mr.SetChains(cm.chains)
		mr.SetModelMax(cm.modelmax)
		mr.SetAtoms(cm.atoms)
		md, err := mr.DoFile()
		if err != nil {
			t.Error(err, "file", testfile)
		}

		valid, invalid := md.NAtomAllChainAll()
		if valid != cm.valid {
			t.Error("valid atoms expected", cm.valid, "got", valid, "n_model", cm.modelmax, "chains", cm.chains)
		}
		if cm.invalid != -1 && invalid != cm.invalid {
			t.Error("invalid atoms expected", cm.invalid, "got", invalid, "n_model", cm.modelmax, "chains", cm.chains)
		}
		fp.Close()
	}
}

var atNameData = []struct {
	chains   []string
	modelmax int16
	atname   string
	nExp     int
}{
	{[]string{}, 1, "xx", 0},
	{[]string{"A"}, 1, "CA", 12},
	{[]string{"A"}, -1, "NZ", 9},
	{[]string{}, -1, "NZ", 9},
	{[]string{}, 1, "NZ", 3},
	{[]string{}, -1, "CQRS", 0},
}

func TestNothingtofind(t *testing.T) {
	type sSlice []string
	var atomlists = []sSlice{
		[]string{"a", "b"},
		[]string{},
		[]string{"CA"},
	}
	testfile := "101d.cif"

	for _, a := range atomlists {
		fp, err := getFp(testdata, testfile, t)
		if err != nil {
			t.Error("Unexpected error file " + testfile)
		}
		mr := NewMmcifReader(fp)
		mr.SetModelMax(1)
		mr.SetAtoms(a)
		if _, err := mr.DoFile(); err != nil {
			t.Error(err, "file", testfile)
		}
		fp.Close()
	}
}

func TestNAtomType(t *testing.T) {
	testfile := "threemodel.cif"

	for _, at := range atNameData {
		fp, err := getFp(testdata, testfile, t)
		if err != nil {
			t.Error("Unexpected error on file " + testfile)
		}
		mr := NewMmcifReader(fp)
		mr.SetChains(at.chains)
		mr.SetModelMax(at.modelmax)
		mr.SetAtoms([]string{})
		md, err := mr.DoFile()
		if err != nil {
			t.Error(err, "file", testfile)
		}
		if got := md.NAtomType(at.atname); got != at.nExp {
			t.Error("expected", at.nExp, "got", got, "at type", at.atname,
				"n_model", at.modelmax, "chains", at.chains)
		}
		fp.Close()
	}
}

// TestAtom4 checks that the atom matching works with a long atom name (4 char)
func TestAtom4(t *testing.T) {
	testfile := "fourletter.cif"
	fp, err := getFp(testdata, testfile, t)
	defer fp.Close()
	if err != nil {
		t.Error("Unexpected error on file " + testfile)
	}
	mr := NewMmcifReader(fp)
	//	mr.SetChains("")
	mr.SetModelMax(-1)
	name := "CQRS"
	mr.SetAtoms([]string{name})
	md, err := mr.DoFile()
	if err != nil {
		t.Error(err, "file", testfile)
	}
	if got := md.NAtomType(name); got != 1 {
		t.Error("Did not find", name, "in file", testfile)
	}
}

// TestChain2 checks that we are happy with chains whose name is two characters
func TestChain2(t *testing.T) {
	testfile := "funny_chain.cif"
	fp, err := getFp(testdata, testfile, t)
	defer fp.Close()
	if err != nil {
		t.Error("Unexpected error on file " + testfile)
	}
	mr := NewMmcifReader(fp)
	mr.SetChains([]string{"Az"})
	mr.SetModelMax(-1)
	name := "CA"
	mr.SetAtoms([]string{name})
	md, err := mr.DoFile()
	if err != nil {
		t.Error(err, "file", testfile)
	}
	expct := 3
	if got := md.NAtomType(name); got != expct {
		t.Error("Expected ", expct, "CAs, got ", got)
	}
}

// Some real files
func TestMmcifReader(t *testing.T) {
	fpdir, err := os.Open(testdata)
	if err != nil {
		t.Error("broke opening testdata directory")
	}
	files, err := fpdir.Readdirnames(0)
	if err != nil {
		t.Error("broke reading file names")
	}
	fpdir.Close()
	for _, f := range files { // change rdr to fp
		rdr, err := getFp(testdata, f, t)
		if err != nil {
			t.Error("file was " + f)
		}
		test1 := "_cell.length_a"
		mr := NewMmcifReader(rdr)
		mr.AddItems([]string{test1})
		mr.SetChains([]string{})
		md, err := mr.DoFile()
		rdr.Close()
		if err != nil {
			t.Error(err.Error() + " on " + f)
		}
		if md == nil {
			t.Error("Nothing found in ", f)
		}
		if md.Data[test1] != "1" {
			t.Error("wrong value or not found: ", test1, "in", f)
		}
	}
}

// Read a file with columns in a non-standard place
func TestFunnyCol(t *testing.T) {
	xyzExpct := [3]float32{27.27, 25.549, -0.624}
	testfile := "simple.cif"
	testat := "O"
	fp, err := getFp(testdata, testfile, t)
	if err != nil {
		t.Error("file was " + testfile)
	}
	defer fp.Close()
	mr := NewMmcifReader(fp)
	mr.SetAtoms([]string{testat})
	md, err := mr.DoFile()
	if err != nil {
		t.Error(err, "file", testfile)
	}
	if xyzS, err := md.GetXyz(0, "A", testat); err != nil {
		t.Error(err)
	} else {
		if XyzNotEq(xyzExpct[:], xyzS[0]) {
			t.Error("Wrong xyz value", testfile)
		}
	}
}

const brokenfiledir string = "testerrors"

func TestErrors(t *testing.T) {
	var files = []struct {
		fname string
		emsg  string
	}{
		{"nosuchfile", "opening file"},
		{"brokeny.cif", "parsing"},
		{"broken_resnum.cif", "parsing"},
		{"rubbish.cif", "Unknown"},
		{"shortline.cif", "Too few"},
		{"inscode.cif", "insertion code length"},
	}
	for _, f := range files {
		var err error
		testfile := f.fname
		fp, err := getFp(brokenfiledir, testfile, t)
		if err == nil {
			mr := NewMmcifReader(fp)
			mr.SetModelMax(1)
			mr.SetAtoms([]string{"CA", "CB"})
			_, err = mr.DoFile()
			fp.Close()
		}
		if err == nil {
			t.Error("should have error on file", testfile)
		}
		if !strings.Contains(err.Error(), f.emsg) {
			t.Error(err, "file", testfile)
		}
	}
}

func TestCmmtscanner(t *testing.T) {
	var ss = []twostring{
		{"some words", "some words"},
		{"#beforecomment#after", ""},
		{"with'quote", "with'quote"},
		{"#hash'#inquote", ""},
		{"hash'#inquote'before#after", "hash'#inquote'before#after"},
		{"ab\"#keep", "ab\"#keep"},
		{"ab\"#keep\"c#", "ab\"#keep\"c#"},
		{"", ""},
	}
	for _, x := range ss {
		scnr := NewCmmtScanner(bytes.NewReader([]byte(x.in)), '#')
		scnr.Cscan()
		b := scnr.Cbytes()
		if string(b) != x.out {
			t.Errorf("Expected\"%s\" got \"%s\"\n", x.out, string(b))
		}
	}
}

type sb []string
type strSlice struct {
	in  string
	out sb
}

func TestSplitCifLine(t *testing.T) {
	var ss = []strSlice{
		{"", sb{""}},
		{"a\"b\"", sb{"a\"b\""}},
		{`b"b"b"b`, sb{"b\"b\"b\"b"}},
		{`b"b"b"b"`, sb{"b\"b\"b\"b\""}},
		{"a b c ", sb{"a", "b", "c"}},
		{"c", sb{"c"}},
		{`aa'aa`, sb{"aa'aa"}},
	}
	scratch := make([][]byte, 3)

	for _, x := range ss {
		tt, err := SplitCifLine([]byte(x.in), scratch)
		if err != nil {
			t.Errorf("Splitting x.in gave error %s\n", err)
		}
		for i, tOut := range tt {
			if string(tOut) != x.out[i] {
				t.Errorf("Splitting <%s> broken, got ", x.in)
				for _, x := range tOut {
					t.Errorf(" <%s>", string(x))
				}
			}
		}
	}
}

func TestSplitCifLine2(t *testing.T) {
	ss := `#This is my test string.
word1 word2
"word1"  	word2
"word1"word2
word1 "word2"
# and a comment in the middle of the file
# and the next should give us errors
   word1 word2

`
	scnr := NewCmmtScanner(bytes.NewReader([]byte(ss)), '#')
	var n_ok, n_broken int
	scratch := make([][]byte, 0)
	for scnr.Cscan() == true && scnr.Cbytes() != nil {
		tt, err := SplitCifLine(scnr.Cbytes(), scratch)
		if err != nil {
			n_broken++
		} else {
			n_ok++
			if len(tt) != 2 {
				t.Errorf("\"%s\" want %d items, got %d", string(scnr.Bytes()), 2, len(tt))
			}
			if string(tt[0]) != "word1" || string(tt[1]) != "word2" {
				t.Errorf("string \"%s\" not broken down correctly", string(scnr.Bytes()))
			}
		}
		scratch = scratch[:0]
	}
	if n_broken != 1 {
		t.Errorf("Expected one error, got %d\n", n_broken)
	}
}

// TestBroken checks that we do get an error on silly strings.
func TestBroken(t *testing.T) {
	ss := []string{
		`'word1'"word2"`,
		`word1 "word2`,
	}
	scratch := make([][]byte, 0)
	for _, s := range ss {
		_, err := SplitCifLine([]byte(s), scratch)
		if err == nil {
			t.Error("Expected an error on string", s)
		}
	}
}

// TestGetChains
func TestGetChains(t *testing.T) {
	testfile := "threemodel.cif"
	fp, err := getFp(testdata, testfile, t)
	if err != nil {
		t.Error("file: " + testfile)
	}
	defer fp.Close()
	mr := NewMmcifReader(fp)
	mr.SetChains([]string{""})
	mr.SetModelMax(3)
	mr.SetAtoms([]string{"CA", "CB", "N"})
	md, err := mr.DoFile()
	if err != nil {
		t.Error(err, "file", testfile)
	}
	if md == nil {
		t.Error("nothing found, md is nil")
	}

	chains := md.GetChains()
	if len(chains) != len(md.Allcoord) {
		t.Error("Problem copying chains")
	}
}

func TestFields(t *testing.T) {
	type ftest struct {
		s string
		a []string
	}
	var tests = []ftest{
		{" 1", []string{"1"}},
		{"", []string{}},
		{" ", []string{}},
		{" 1", []string{"1"}},
		{"1", []string{"1"}},
		{" 1 ", []string{"1"}},
		{"1 2", []string{"1", "2"}},
		{" 1 2", []string{"1", "2"}},
		{"1   2", []string{"1", "2"}},
		{"1   2 ", []string{"1", "2"}},
		{"1   2    ", []string{"1", "2"}},
		{"12 34", []string{"12", "34"}},
		{"12 34 ", []string{"12", "34"}},
		{"1 2 3 4", []string{"1", "2", "3", "4"}},
		{"1 2 3 4 ", []string{"1", "2", "3", "4"}},
		{"1  2 3 4 ", []string{"1", "2", "3", "4"}},
		{"ATOM 1805  O O    . GLY A 1 10 ? -16.616 0.276   -4.686  1.00 0.00 ?  299 GLY A O    2",
			[]string{"ATOM", "1805", "O", "O", ".", "GLY", "A", "1", "10", "?", "-16.616", "0.276", "-4.686", "1.00", "0.00", "?", "299", "GLY", "A", "O", "2"}},
	}

	for _, tt := range tests {
		var scrtch [40]BSlice
		ret := Fields([]byte(tt.s), scrtch[:])
		if len(ret) != len(tt.a) {
			t.Errorf("Wanted %d fields, got %d, string '%s'", len(tt.a), len(ret), tt.s)
		}
		for i, a := range tt.a {
			if string(ret[i]) != a {
				t.Errorf("fields mismatch want '%s' got '%s'", tt.a[i], ret[i])
			}
		}
	}
}

func TestResolution(t *testing.T) {
	testfile := "resolution.cif"
	test1 := []string{"_refine.ls_d_res_high", "_reflns.d_resolution_high"}
	fp, err := getFp(testdata, testfile, t)
	defer fp.Close()
	if err != nil {
		t.Error("Unexpected error on file " + testfile)
	}
	mr := NewMmcifReader(fp)
	mr.AddItems (test1)
	md, err := mr.DoFile()

	if err != nil {
		t.Error(err.Error() + " on " + testfile)
	}
	if md == nil {
		t.Error("Nothing found in ", testfile)
	}
	found := false
	for _, s := range test1 {
		if _, ok := md.Data[s]; ok == true {
			found = true
		}
	}
	if !found { t.Error ("did not find resolution") }
}


func TestFieldsLong(t *testing.T) {
	const small = 5
	var scrtch [small]BSlice
	in := BSlice(" 1 2 3 4 5 6 7 8 9 0 ")
	out := Fields(in, scrtch[:])
	if len(out) != small {
		t.Error("Problem when scratch array is too small")
	}
}
