// Feb 2018

// Package gotoh implements the Gotoh version of pair-wise alignments.
// We use a full scoring matrix, and use this during the summation.
// It will change, so we can declare an "aligner" with its various
// temporary storage. This can then be re-used over different
// alignments and will go away when the aligner goes away.
package gotoh

import (
	"fmt"
	"github.com/andrew-torda/goutil/matrix"
)

// Al_type is a byte which can be global or local. By making it its
// own type, we can add a String() method to it.
type Al_type byte

// Local/Global are exported constants to say what kind of alignment one wants.
const (
	Local  Al_type = iota // Local alignment
	Global                // global alignment
)

// Pnlty has the gap opening and widening values. Note, this is different to
// some earlier code. Opening costs you -(Open+Wdn). Each extension costs
// -Wdn
type Pnlty struct {
	Open float32
	Wdn  float32
}

type Match_scr struct {
	Match    float32 // matched characters
	Mismatch float32 // mismatched
}

// Match_score is for telling a score routine how to score an alignment.
// scoring
type Al_score struct {
	Pnlty           // open and widen penalties
	Al_type Al_type // local / global
}

const (
	diag byte = iota // diagonal movement
	pway             // along the P direction, vertical, over rows
	qway             // Q direction, horizontal, over columns
	stop             // Can be used to signal traceback should stop
)

type ipair struct {
	i, j int // to store aligned pairs
}

const bigf float32 = -1e+38

// String for the alignment type is mainly for debugging.
func (a Al_type) String() string {
	if a == Local {
		return "local"
	}
	return "global"
}

// IdentScore fills out a score matrix using identity. Values for match/mismatch
// come from the scr structure.
// for an M x N pair, we have an M x N matrix. There is no extra room
// at the start and end.
func IdentScore(s []byte, t []byte, scr *Match_scr) (smat *matrix.FMatrix2d) {
	smat = matrix.NewFMatrix2d(len(s), len(t))
	mat := smat.Mat
	for i, cs := range s {
		for j, ct := range t {
			if cs == ct {
				mat[i][j] = scr.Match
			} else {
				mat[i][j] = scr.Mismatch
			}
		}
	}
	return
}

// PrintSeqDebug is a primitive printer for aligned sequences, but
// it is essential for debugging.
func PrintSeqDebug(verbose bool, pairlist []ipair, s, t []byte, al_type Al_type) {
	if !verbose {
		return
	}
	var outs1, outs2 string
	for _, p := range pairlist {
		if p.i == -1 {
			outs1 = outs1 + "-"
		} else {
			outs1 = outs1 + string(s[p.i])
		}
		if p.j == -1 {
			outs2 += "-"
		} else {
			outs2 += string(t[p.j])
		}
	}
	fmt.Println("aligned ", al_type, ":\n", outs1, "\n", outs2)
}

// traceback gotoh
// Return the list of pairs in the pairlist and the maximum total score
// dir is the matrix with directions. scr_mat is the original score matrix,
// but we will use it to hold the summations. al_type can be local or
// global
func traceback(dir *[][]byte, scr_mat [][]float32, al_type Al_type) (
	pairlist []ipair, max_scr float32) {
	nr := len(scr_mat)
	nc := len(scr_mat[0])
	max_scr = scr_mat[nr-1][nc-1]
	max_i, max_j := nr-1, nc-1
	{
		bigger := nr     // Take a guess as to how much space we
		if nc > bigger { // might need for saving the aligned pairs.
			bigger = nc // Use the longer string and add 10 %.
		}
		pairlist = make([]ipair, 0, bigger+bigger/10)
	}
	if al_type == Local { //             Local alignments start from
		for i, row := range scr_mat { // the highest score, even if it
			for j := range row { //      is not at one of the edges
				if scr_mat[i][j] > max_scr {
					max_scr = scr_mat[i][j]
					max_i, max_j = i, j
				}
			}
		}
	} else {
		for i, col := 0, nc-1; i < nr; i++ { // Look in last column
			if scr_mat[i][col] > max_scr { //   for highest score
				max_scr = scr_mat[i][col]
				max_i, max_j = i, col
			}
		}
		for j, row := 0, nr-1; j < nc; j++ { // Look in last row
			if scr_mat[row][j] > max_scr { //   for highest score
				max_scr = scr_mat[row][j]
				max_i, max_j = row, j
			}
		}
	}

	if al_type == Local { // This seems to work well
		const thresh = 0
		var i, j int
		for i, j = max_i, max_j; (*dir)[i][j] != stop && scr_mat[i][j] > thresh; {
			switch (*dir)[i][j] {
			case diag:
				pairlist = append(pairlist, ipair{i, j})
				i--
				j--
			case pway:
				pairlist = append(pairlist, ipair{i, -1})
				i--
			case qway:
				pairlist = append(pairlist, ipair{-1, j})
				j--
			}
		}
		if scr_mat[i][j] > thresh {
			pairlist = append(pairlist, ipair{i, j})
		}
	} else {
		if max_i == nr-1 {
			for jj := nc - 1; jj > max_j; jj-- {
				pairlist = append(pairlist, ipair{-1, jj})
			}
		} else if max_j == nc-1 {
			for ii := nr - 1; ii > max_i; ii-- {
				pairlist = append(pairlist, ipair{ii, -1})
			}
		}

		const thresh = bigf
		var i, j int
		for i, j = max_i, max_j; (*dir)[i][j] != stop; {
			switch (*dir)[i][j] {
			case diag:
				pairlist = append(pairlist, ipair{i, j})
				i--
				j--
			case pway:
				pairlist = append(pairlist, ipair{i, -1})
				i--
			case qway:
				pairlist = append(pairlist, ipair{-1, j})
				j--
			}
		}
		pairlist = append(pairlist, ipair{i, j})
		i--
		for ; i >= 0; i-- {
			pairlist = append(pairlist, ipair{i, -1})
		}
		j--
		for ; j >= 0; j-- {
			pairlist = append(pairlist, ipair{-1, j})
		}
	}

	for i, j := 0, len(pairlist)-1; i < j; i, j = i+1, j-1 {
		pairlist[i], pairlist[j] = pairlist[j], pairlist[i]
	}

	return pairlist, max_scr
}

// Align implements Gotoh, O. J. Mol. Biol. (1982) 162, 705-708.
// It does not have the bugs described in Flouri, T, Kobert, K., Rognes, T
// and Stamatakis, doi: http://dx.doi.org/10.1101/031500 (2015).
func Align(scr_mat_mat *matrix.FMatrix2d, scr_scheme *Al_score) (
	pairlist []ipair, max_scr float32) {
	var max = func(a, b float32) float32 {
		if a > b {
			return a
		}
		return b
	}

	opn := -scr_scheme.Open
	wdn := -scr_scheme.Wdn
	w1 := opn - scr_scheme.Wdn
	scr_mat := scr_mat_mat.Mat
	nrow, ncol := len(scr_mat), len(scr_mat[0])
	if nrow < 1 || ncol < 1 {
		return
	}
	dirtmp := (matrix.NewBMatrix2d(nrow, ncol)) // Where we store directions
	dir := dirtmp.Mat                           // for the traceback

	for _, c := range dir {
		c[0] = stop
	}
	for i := range dir[0] {
		dir[0][i] = stop
	}

	p := make([]float32, ncol)

	if scr_scheme.Al_type == Local { //  Start and column can not be
		for _, row := range scr_mat { // negative in a local alignment
			row[0] = max(row[0], 0)
		}
		for i := range scr_mat[0] {
			scr_mat[0][i] = max(scr_mat[0][i], 0)
		}
	}
	//	scr_mat[0][0] = max (scr_mat[0][0], w1)
	for i, qprev := 1, bigf; i < ncol; i++ { //  special case first row
		q := max(scr_mat[0][i-1]+w1, qprev+wdn)
		if q >= scr_mat[0][i] {
			scr_mat[0][i] = q
			dir[0][i] = qway
		}
		qprev = q
	}

	for i, qprev := 1, bigf; i < nrow; i++ { // special case first column
		q := max(scr_mat[i-1][0]+w1, qprev+wdn)
		if q >= scr_mat[i][0] {
			scr_mat[i][0] = q
			dir[i][0] = pway
		}
		qprev = q
	}
	for i := range p {
		p[i] = bigf
	}

	for i := 1; i < nrow; i++ { // Indexing is such that we walk
		qprev := bigf //     along each row, left to right.
		for j := 1; j < ncol; j++ {
			best := scr_mat[i][j] + scr_mat[i-1][j-1]
			drctn := diag
			p[j] = max(scr_mat[i-1][j]+w1, p[j]+wdn)
			q := max(scr_mat[i][j-1]+w1, qprev+wdn)
			if p[j] > best {
				best, drctn = p[j], pway
			}
			if q > best {
				best, drctn = q, qway
			}
			scr_mat[i][j] = best
			dir[i][j] = byte(drctn)
			qprev = q
		}
	}
	pairlist, max_scr = traceback(&dir, scr_mat, scr_scheme.Al_type)
	return pairlist, max_scr
}
