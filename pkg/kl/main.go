// 28 may 2020

package kl

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/andrew-torda/goutil/seq"
	"github.com/andrew-torda/matrix"
)

// CmdFlag is literally command line flags after parsing
type CmdFlag struct {
	Offset      int  // Add this to the residue numbering on output
	GapsAreChar bool // Do we keep gaps ? Are gaps a valid symbol ?
	NSym        int  // Set the number of symbols in sequences
}

// seqX are the elements of a SeqGrp structure which are
// relevant to the calculations for Kullbach-Leibler divergence.
// We copy the slices we need and can free up things like the original
// sequences. It is only exported so we can use it in testing.
type SeqX struct {
	entropy []float32
	counts  *matrix.FMatrix2d
	revmap  []uint8
	nseq    int
	logbase int
}

func (seqx *SeqX) GetLen() int { return len(seqx.entropy) }

// seqX gets the relevant information for KL calculation from a sequence
// group. It only goes into its own function so it can be called
// during testing.
func getSeqX(seqgrp *seq.SeqGrp, seqX *SeqX, flags *CmdFlag) error {
	var err error
	seqX.entropy, err = seqgrp.Entropy(flags.GapsAreChar)
	if err != nil {
		return err
	}
	var gapsAreChars = true

	logbase, err := seqgrp.GetLogBase(gapsAreChars)
	if err != nil {
		return err
	}
	seqX.counts = seqgrp.GetCounts()
	seqX.revmap = seqgrp.GetRevmap()
	seqX.nseq = seqgrp.GetNSeq()
	seqX.logbase = logbase
	return nil
}

// getent gets the entropy in a file. It will be called once for each of
// the two files.
// The goroutine runs for each of the two files at the same time. The
// background process gets a wait group, so we can signal that he is finished.
// The foreground process gets a nil pointer.
func getent(flags *CmdFlag, infile string, seqX *SeqX, err *error,
	wg *sync.WaitGroup, symSync *seq.SymSync) {
	bailout := func() {
		var junk [seq.MaxSym]bool
		symSync.UChan <- junk
	}
	if wg != nil { // Only use the wait group for the
		defer wg.Done() // background goroutine.
	}
	s_opts := &seq.Options{
		Vbsty: 0, Keep_gaps_rd: true,
		Dry_run:      true,
		Rmv_gaps_wrt: true,
	}

	seqgrp, _, e := seq.Readfile(infile, s_opts)
	if e != nil {
		*err = fmt.Errorf("Fail reading sequences: %w", e)
		bailout()
		return //
	}
	if seqgrp.GetNSeq() == 0 {
		*err = errors.New("Zero sequences found in file " + infile)
		bailout()
		return //
	}
	seqgrp.Upper()

	seqgrp.SetSymUsed(symSync)
	*err = getSeqX(&seqgrp, seqX, flags)
}

// readtwofiles reads the two input sequence files. One of them will
// be read in the background. We use a waitgroup for synchronising.
func readtwofiles(flags *CmdFlag, file1, file2 string,
	seqXP, seqXQ *SeqX) error {
	var err1, err2 error
	var wg sync.WaitGroup
	var once sync.Once
	symSync := seq.SymSync{Once: once, UChan: make(chan [seq.MaxSym]bool)}
	wg.Add(1)
	go getent(flags, file1, seqXP, &err1, &wg, &symSync)
	getent(flags, file2, seqXQ, &err2, nil, &symSync)
	wg.Wait()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

// kl calculates the kullbach-leibler distance
// When in doubt think
//  d(p|q) = sum(p_i log (p_i/q_i))
//         = sum ( p_i (log(p_i) - log(q_i)))
//         = sum (p_i (log(p_i)) - p_i log(q_i))    , first term is just entropy of p
//         = s_p - p_i log(q_i)         and this is what we program up.
//                 ^^^^^^^^^^^ This is called "cross" in the code below
// When one of the distributions goes to zero, divergence goes to
// infinity. Let us try either setting it to a max value or better...
// We have N sequences for distribution q. We say the frequency is less
// than 1/ N. This means we calculate log(1 / (N+1) and use it for missing values.
// Need ent_p, p, q
type KL_in struct {
	ent_p    []float32   // entropy for p distribution
	counts_p [][]float32 // counts/frequencies for p distribution
	counts_q [][]float32 // " "        "   "    "  q distribution
	num_q    int         // number of sequences in q distribution
	logbase  int         // base to use for logarithms
}
func breaker(){}
func kl(kl_in *KL_in, kl []float32) {
	logbase := math.Log(float64(kl_in.logbase))
	logb := func(x float32) float32 { // log will be base 4 or 20
		return float32(math.Log(float64(x)) / logbase)
	}
	pLnPQ := make([]float32, len(kl_in.ent_p)) // p log (q) term
	one_num_q := 1.0 / float32(kl_in.num_q+1)
	log_one_num_q := logb(one_num_q)
	seqLen := len(kl_in.counts_p[0])

	for icol := 0; icol < seqLen; icol++ { //      i position in sequence
		if (icol >= 3) {breaker()}
		for irow := 0; irow < kl_in.logbase; irow++ { //  j is symbol
			if kl_in.counts_p[irow][icol] == 0 {
				continue
			}
			var log_q float32
			if kl_in.counts_q[irow][icol] == 0 {
				log_q = log_one_num_q
			} else {
				log_q = logb(kl_in.counts_q[irow][icol])
			}
			pLnPQ[icol] += kl_in.counts_p[irow][icol] * log_q
		}
	} // pLnPQ now holds the second term in the formula
	for i, ep := range kl_in.ent_p {
		kl[i] = -ep - pLnPQ[i]
	}
}

// innerCosProd does the cosine product of two vectors
func innerCosProd(p, q []float32) float32 {
	var p_sq, q_sq, res float64
	for _, x := range p {
		p_sq += float64(x * x)
	}
	for _, x := range q {
		q_sq += float64(x * x)
	}
	p_sq = math.Sqrt(p_sq)
	q_sq = math.Sqrt(q_sq)
	for i := range p {
		pp := float64(p[i]) / p_sq
		qq := float64(q[i]) / q_sq
		res += pp * qq
	}
	return float32(res)
}

// calcCosProd calculates the cosine product of two distributions
// Walk down each column in p. Sum the squares. Take the square root.
// Walk down each column and divide by the square root.
// Repeat for the second vector.
// Return the dot product.
func calcCosSim(counts_p [][]float32, counts_q [][]float32, cosSim []float32) {
	if len(counts_p) != len(counts_q) {
		panic("silly programming bug in calccosprod")
	}
	pvec := make([]float32, len(counts_p))
	qvec := make([]float32, len(counts_p))
	for i := range counts_p { // i is position in sequence
		for j := 0; j < len(counts_p); j++ {
			pvec[j] = counts_p[i][j]
			qvec[j] = counts_q[i][j]
		}
		r := innerCosProd(pvec, qvec)
		cosSim[i] = r
	}
}

// klFromSeqX takes a pair of seqXs and returns the KL
// distance in one direction. To get the other direction, call with
// the arguments reversed.
func klFromSeqX(seqXP, seqXQ *SeqX, klP []float32) {
	klIn := KL_in{
		ent_p:    seqXQ.entropy,
		counts_p: seqXP.counts.Mat,
		counts_q: seqXQ.counts.Mat,
		num_q:    seqXQ.nseq,
		logbase:  seqXP.logbase,
	}
	kl(&klIn, klP)
}

// Mymain is the main function for kullback-leibler distance
func Mymain(flags *CmdFlag, fileP, fileQ, outfile string) (err error) {
	var seqXP, seqXQ SeqX

	const mismatch = "Sequence length mismatch. %s: len %d, %s: len %d"
	if err := readtwofiles(flags, fileP, fileQ, &seqXP, &seqXQ); err != nil {
		return err
	} // Maybe take the next few lines and bundle them up into sanitycheck()
	lenAln := seqXP.GetLen() // Check alignments have same length.
	if l2 := seqXQ.GetLen(); l2 != lenAln {
		return fmt.Errorf(mismatch, fileP, lenAln, fileQ, l2)
	}
	// later, lets change this to one allocation and make three pieces
	cosSim := make([]float32, seqXP.GetLen())
	klP := make([]float32, seqXP.GetLen())
	klQ := make([]float32, seqXP.GetLen())
	// We can do the next three calculations at the same time
	klFromSeqX(&seqXP, &seqXQ, klP)
	klFromSeqX(&seqXQ, &seqXP, klQ)
	calcCosSim(seqXP.counts.Mat, seqXQ.counts.Mat, cosSim)
	return nil
}