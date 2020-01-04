// Get the tests from the greek paper.
// Includes the example from the Altschul paper which cannot be
// done with the method as described in the original paper.

package gotoh_test

import (
	gth "andrew/gotoh"
	"andrew/randseq"
	"andrew/seq"
	"andrew/submat"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

const local, global gth.Al_type = gth.Local, gth.Global

func rev(s string) string {
	t := []byte(s)
	for i, j := 0, len(t)-1; i < j; i, j = i+1, j-1 {
		t[i], t[j] = t[j], t[i]
	}
	return string(t)
}

func TestRand(t *testing.T) {
	var err error
	s1 := randseq.New(seq.Protein, 10)
	s2 := make([]byte, len(s1))
	copy(s2, s1)
	verbose := true
	n_change := randseq.Mutate(seq.Protein, 1./3., s2)
	n_to_del := int(float32(len(s1)) * 0.2)
	if s2, err = randseq.DelN(n_to_del, s2); err != nil {
		panic("deleting " + string(n_to_del))
	}

	matches := gth.Match_scr{5, -1}
	scr_mat := gth.IdentScore(s1, s2, &matches)
	al_score := gth.Al_score{gth.Pnlty{2, 1}, global}
	pairlist, max_scr := gth.Align(scr_mat, &al_score)
	if testing.Verbose() {
		fmt.Println("at mutation step, changed", n_change, "sites.")
		gth.PrintSeqDebug(verbose, pairlist, s1, s2, global)
		fmt.Println("score", max_scr)
	}
}

var testpairs = []struct {
	s1      string        // I tried to line this up and make it readable.
	s2      string        // gofmt removes all the excess spaces
	m_scr   gth.Match_scr // Given to identity score function
	a_scr   gth.Pnlty     // Gap open and widen penalties
	scr_exp []float32     // expected scores, local and global alignments
}{ // s1  ,  s2,                       {match, mismatch, {open, widen}}, al_type, expected scores
	{"bcde", "ae", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{5, 3}},
	{"abcdefghi", "bcgi", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{14, 14}},
	{"abcdefg", "aceh", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{11, 9}},
	{"ae", "abcd", gth.Match_scr{5, -9}, gth.Pnlty{1, 1}, []float32{5, 3}},
	{"aceh", "abcdefxy", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{11, 9}},
	{"aceh", "abcdefxyz", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{11, 9}},
	{"exz", "abcdefxyz", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{11, 11}},
	{"dxz", "abcdefxyz", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{10, 10}},
	{"abcde", "abe", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{12, 12}},
	{"abcdef", "abde", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{18, 18}},
	{"aceg", "abcdef", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{11, 9}},
	//	{"a", "b",        gth.Match_scr{5, -1}, gth.Pnlty{1, 1}, []float32{-1, -1}},
	{"abcde", "bcd", gth.Match_scr{2, 1}, gth.Pnlty{2, 4}, []float32{6, 6}},
	{"a", "a", gth.Match_scr{5, 2}, gth.Pnlty{1, 1}, []float32{5, 5}},
	{"abc", "xaby", gth.Match_scr{5, -1}, gth.Pnlty{1, 1}, []float32{10, 9}},
	{"abcd", "abd", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{13, 13}},
	{"abcdef", "abf", gth.Match_scr{5, -2}, gth.Pnlty{1, 1}, []float32{11, 11}},
	{"xabc", "aby", gth.Match_scr{5, -1}, gth.Pnlty{1, 1}, []float32{10, 9}},
	// From the Altschul paper...
	{"AAAGGG", "TTAAAAGGGGTT", gth.Match_scr{1, -1}, gth.Pnlty{5, 1}, []float32{6, 6}},
}

func TestGotoh(t *testing.T) {
	const verbose = false
	var vprint = func(verbose bool, a ...interface{}) {
		if verbose {
			fmt.Println(a...)
		}
	}

	var f = func(s1, s2 string, m_scr *gth.Match_scr, a_scr *gth.Al_score) float32 {
		vprint(verbose, "%10s %10s %v\n", s1, s2, m_scr, a_scr)
		scr_mat := gth.IdentScore([]byte(s1), []byte(s2), m_scr)
		pairlist, max_scr := gth.Align(scr_mat, a_scr)
		scr_mat = nil
		gth.PrintSeqDebug(verbose, pairlist, []byte(s1), []byte(s2), a_scr.Al_type)
		vprint(verbose, "---------------------------------")
		return max_scr
	}
	var lg = []gth.Al_type{global, local}
	for _, typ := range lg {
		for _, x := range testpairs {
			s1 := x.s1
			s2 := x.s2
			//			x.a_scr.Al_score.Al_type = typ // type of alignment, local or global
			tmp := gth.Al_score{x.a_scr, typ}
			scr_1 := f(s1, s2, &x.m_scr, &tmp)
			scr_2 := f(s2, s1, &x.m_scr, &tmp)
			scr_3 := f(rev(s2), rev(s1), &x.m_scr, &tmp)
			if scr_1 != scr_2 {
				t.Fatal("string1/string2 string2/string1 scores not equal.\n",
					"Strings were ", s1, s2, "scores", scr_1, scr_2, "expected", x.scr_exp[typ])
			}
			if scr_2 != scr_3 && len(s1) > 2 && len(s2) > 2 {
				t.Fatal("string3/string2 scores not equal with reversed strings\n",
					"Strings were ", s1, s2, typ)
			}

			exp_scr := x.scr_exp[typ]
			if scr_1 != exp_scr {
				t.Fatal("alignment type", tmp.Al_type,
					"Wrong score while aligning\n", s1, "and", s2, "Expected", exp_scr, "got", scr_1)
			}
		}
	}
}

func make_submat() (subst_mat *submat.Submat, err error) {
	matrix_text := []byte(`#  Matrix made by matblas from blosum62.iij
#  * column uses minimum score
#  BLOSUM Clustered Scoring Matrix in 1/2 Bit Units
#  Blocks Database = /data/blocks_5.0/blocks.dat
#  Cluster Percentage: >= 62
#  Entropy =   0.6979, Expected =  -0.5209
   A  R  N  D  C  Q  E  G  H  I  L  K  M  F  P  S  T  W  Y  V  B  Z  X  *
A  4 -1 -2 -2  0 -1 -1  0 -2 -1 -1 -1 -1 -2 -1  1  0 -3 -2  0 -2 -1  0 -4 
R -1  5  0 -2 -3  1  0 -2  0 -3 -2  2 -1 -3 -2 -1 -1 -3 -2 -3 -1  0 -1 -4 
N -2  0  6  1 -3  0  0  0  1 -3 -3  0 -2 -3 -2  1  0 -4 -2 -3  3  0 -1 -4 
D -2 -2  1  6 -3  0  2 -1 -1 -3 -4 -1 -3 -3 -1  0 -1 -4 -3 -3  4  1 -1 -4 
C  0 -3 -3 -3  9 -3 -4 -3 -3 -1 -1 -3 -1 -2 -3 -1 -1 -2 -2 -1 -3 -3 -2 -4 
Q -1  1  0  0 -3  5  2 -2  0 -3 -2  1  0 -3 -1  0 -1 -2 -1 -2  0  3 -1 -4 
E -1  0  0  2 -4  2  5 -2  0 -3 -3  1 -2 -3 -1  0 -1 -3 -2 -2  1  4 -1 -4 
G  0 -2  0 -1 -3 -2 -2  6 -2 -4 -4 -2 -3 -3 -2  0 -2 -2 -3 -3 -1 -2 -1 -4 
H -2  0  1 -1 -3  0  0 -2  8 -3 -3 -1 -2 -1 -2 -1 -2 -2  2 -3  0  0 -1 -4 
I -1 -3 -3 -3 -1 -3 -3 -4 -3  4  2 -3  1  0 -3 -2 -1 -3 -1  3 -3 -3 -1 -4 
L -1 -2 -3 -4 -1 -2 -3 -4 -3  2  4 -2  2  0 -3 -2 -1 -2 -1  1 -4 -3 -1 -4 
K -1  2  0 -1 -3  1  1 -2 -1 -3 -2  5 -1 -3 -1  0 -1 -3 -2 -2  0  1 -1 -4 
M -1 -1 -2 -3 -1  0 -2 -3 -2  1  2 -1  5  0 -2 -1 -1 -1 -1  1 -3 -1 -1 -4 
F -2 -3 -3 -3 -2 -3 -3 -3 -1  0  0 -3  0  6 -4 -2 -2  1  3 -1 -3 -3 -1 -4 
P -1 -2 -2 -1 -3 -1 -1 -2 -2 -3 -3 -1 -2 -4  7 -1 -1 -4 -3 -2 -2 -1 -2 -4 
S  1 -1  1  0 -1  0  0  0 -1 -2 -2  0 -1 -2 -1  4  1 -3 -2 -2  0  0  0 -4 
T  0 -1  0 -1 -1 -1 -1 -2 -2 -1 -1 -1 -1 -2 -1  1  5 -2 -2  0 -1 -1  0 -4 
W -3 -3 -4 -4 -2 -2 -3 -2 -2 -3 -2 -3 -1  1 -4 -3 -2 11  2 -3 -4 -3 -2 -4 
Y -2 -2 -2 -3 -2 -1 -2 -3  2 -1 -1 -2 -1  3 -3 -2 -2  2  7 -1 -3 -2 -1 -4 
V  0 -3 -3 -3 -1 -2 -2 -3 -3  3  1 -2  1 -1 -2 -2  0 -3 -1  4 -3 -2 -1 -4 
B -2 -1  3  4 -3  0  1 -1  0 -3 -4  0 -3 -3 -2  0 -1 -4 -3 -3  4  1 -1 -4 
Z -1  0  0  1 -3  3  4 -2  0 -3 -3  1 -1 -3 -1  0 -1 -3 -2 -2  1  4 -1 -4 
X  0 -1 -1 -1 -2 -1 -1 -1 -1 -1 -1 -1 -1 -1 -2  0  0 -2 -1 -1 -1 -1 -1 -4 
* -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4 -4  1 `)
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		return
	}
	defer os.Remove(tmpfile.Name()) // clean up
	fname := tmpfile.Name()
	if _, err = tmpfile.Write(matrix_text); err != nil {
		return
	}
	tmpfile.Close()
	if subst_mat, err = submat.Read(fname); err != nil {
		return
	}
	return subst_mat, err
}

// TestWithBlosum
// You have to look at score in detail to see if things are working,
// but what we test here is that the relevant calls and return
// types are still up to date.
func TestWithBlosum(t *testing.T) {
	seqs := []string{"acdefgacdefg", "cdefgacsfg", "cdefgactg", "cdefgacwg"}
	pnltys := gth.Pnlty{2, 2}
	subst_mat, err := make_submat()
	if err != nil {
		fmt.Print(err)
	}
	al_details := gth.Al_score{
		pnltys,
		gth.Local,
	}
	for i, s := range seqs {
		for j := i + 1; j < len(seqs); j++ {
			t := seqs[j]
			scr_mat := subst_mat.ScoreSeqs([]byte(s), []byte(t))
			pairlist, a_scr := gth.Align(scr_mat, &al_details)
			if testing.Verbose() {
				gth.PrintSeqDebug(true, pairlist, []byte(s), []byte(t), gth.Global)
				fmt.Println("score", a_scr)
			}
		}
	}
}
