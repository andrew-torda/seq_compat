package altschul

// package altschul implements Altschul, SF and Erickson, BW,
// Bull Math Biol, 48, 603-616 1986

import (
	"github.com/goutil/matrix"
	"fmt"
	"os"
)

type pnlty struct {
	open float32
	wdn  float32
}

type match_score struct {
	match    float32
	mismatch float32
	pnlty
}

const (
	a_ = 1 << iota // uses V (vertical)
	b_             // uses H (horizonta)
	c_             // " "  D (diagonal)
	d_             // amongst i+1,j through N_i, one uses V_i,j
	e_             // amongst
	f_
	g_
)

func identScore(s, t []byte, scr match_score) (smat *matrix.FMatrix2d) {
	smat = matrix.NewFMatrix2d(len(s), len(t))
	for i, cs := range s {
		for j, ct := range t {
			if cs == ct {
				smat.Mat[i][j] = scr.match
			} else {
				smat.Mat[i][j] = scr.mismatch
			}
		}
	}
	return smat
}

// altschul does wondrous things
// I think the strategy is to get it working using full arrays.
// Then "In place of arrays P and R two one-dimensional number
// arrays, and in place of array Q a variable, provide sufficient storage"
// index i ranges from 0 to M and index j from 0 to N.
func Altschul(s, t []byte) {
	var min = func(a, b float32) float32 {
		if a < b {
			return a
		}
		return b
	}
	var max = func(a, b float32) float32 {
		if a > b {
			return a
		}
		return b
	}
	var score = match_score{match: 1, mismatch: 1, pnlty: pnlty{1, 1}}
	var bigf float32 = 2*score.open + max(float32(len(s)), float32(len(t))) + 1
	d := (matrix.NewBMatrix2d(len(s)+2, len(t)+2)).Mat
	p := (matrix.NewFMatrix2d(len(s)+1, len(t)+1)).Mat
	q := (matrix.NewFMatrix2d(len(s)+1, len(t)+1)).Mat
	r := (matrix.NewFMatrix2d(len(s)+1, len(t)+1)).Mat

	for j := 0; j < len(t)+1; j++ {
		p[0][j] = bigf
		r[0][j] = score.open + float32(j)*score.wdn
	}
	for i := 0; i < len(s)+1; i++ {
		q[i][0] = bigf
		r[i][0] = score.open + float32(i)*score.wdn
	}
	r[0][0] = 0.0
	d[len(s)][len(t)] = c_
	for i := 1; i < len(s); i++ {
		for j := 1; j < len(t); j++ {
			p[i][j] = score.wdn + min(p[i-1][j], r[i-1][j]+score.open)
		}
	}
	fmt.Fprintln(os.Stdout, d)
}

func runme() {
	//	fmt.Fprint (ioutil.Discard, runtime.Version)// get rid of this later
	s := "abcde"
	t := "bc"
	Altschul([]byte(s), []byte(t))
}
