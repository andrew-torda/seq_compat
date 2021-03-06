// 28 may 2020

package kl

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sync"

	"github.com/andrew-torda/matrix"
	"github.com/andrew-torda/seq_compat/pkg/seq"
	"github.com/andrew-torda/seq_compat/pkg/seq/common"
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
	counts     *matrix.FMatrix2d
	revmap     []uint8
	nseq       int
	len        int // Sequence length
	logbase    int
	gapMapping uint8 // which row is the gap character
}

func (seqx *SeqX) GetLen() int { return seqx.len }

// seqX gets the relevant information for KL calculation from a sequence
// group. It only goes into its own function so it can be called
// during testing.
func extractSeqX(seqgrp *seq.SeqGrp, seqX *SeqX, flags *CmdFlag) error {
	var err error
	var gapsAreChars = false

	logbase := seqgrp.GetLogBase(gapsAreChars)
	if err != nil {
		return err
	}
	seqgrp.UsageFrac(gapsAreChars)
	seqX.len = seqgrp.GetLen()
	seqX.counts = seqgrp.GetCounts()
	seqX.revmap = seqgrp.GetRevmap()
	seqX.nseq = seqgrp.NSeq()
	seqX.logbase = logbase
	seqX.gapMapping = seqgrp.GetMapping(common.GapChar)
	return nil
}

// getseqX goes to a file name, extracts the information we want to keep, rather
// than storing the full sequences. It will be called on each input file,
// in parallel. The first call is running in the background, so if he gets a
// non-zero waitgroup, he knows to call wg.Done().
// The foreground process gets a nil pointer.
func getseqX(wg *sync.WaitGroup, flags *CmdFlag, infile string, seqX *SeqX,
	err *error, frmMrgChn chan [seq.MaxSym]bool, toMrgChn chan [seq.MaxSym]bool) {
	if wg != nil {
		defer wg.Done()
	}
	bailout := func() {
		var junk [seq.MaxSym]bool
		toMrgChn <- junk
		<-frmMrgChn
	}

	s_opts := &seq.Options{}

	seqgrp, e := seq.Readfile(infile, s_opts)
	if e != nil {
		*err = fmt.Errorf("Fail reading sequences: %w", e)
		bailout()
		return
	}
	if seqgrp.NSeq() == 0 {
		*err = errors.New("Zero sequences found in file " + infile)
		bailout()
		return
	}
	seqgrp.Upper()
	seqgrp.SetSymUsedWithChan(frmMrgChn, toMrgChn)
	*err = extractSeqX(seqgrp, seqX, flags)
}

// mergelists merges two lists of symbols that have been
// used. It reads each list from a channel, merges them
// and returns the merged list, which will have overwritten
// the first list it received.
func mergelists(wg *sync.WaitGroup, frmMrgChn chan [seq.MaxSym]bool, toMrgChn chan [seq.MaxSym]bool) {
	a1, a2 := <-toMrgChn, <-toMrgChn
	for i := range a1 {
		a1[i] = a1[i] || a2[i]
	}

	frmMrgChn <- a1
	frmMrgChn <- a1
	close(frmMrgChn)
	wg.Done()
}

// readtwofiles reads the two input sequence files. One of them will
// be read in the background. We use a waitgroup for synchronising.
func readtwofiles(flags *CmdFlag, file1, file2 string, seqXP, seqXQ *SeqX) error {
	var err1, err2 error
	var wg sync.WaitGroup

	frmMrgChn := make(chan [seq.MaxSym]bool)
	toMrgChn := make(chan [seq.MaxSym]bool)
	wg.Add(1)
	go mergelists(&wg, frmMrgChn, toMrgChn)
	wg.Add(1)
	go getseqX(&wg, flags, file1, seqXP, &err1, frmMrgChn, toMrgChn)
	getseqX(nil, flags, file2, seqXQ, &err2, frmMrgChn, toMrgChn)
	wg.Wait()
	close(toMrgChn)
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

// kl_in just tames the set of arguments we will need for the kl function.
type klIn struct {
	counts_p [][]float32 // counts/frequencies for p distribution
	counts_q [][]float32 // " "        "   "    "  q distribution
	num_q    int         // number of sequences in q distribution
	logbase  int         // base to use for logarithms
}

// kl calculates the kullbach-leibler distance
// When one of the distributions goes to zero, divergence goes to
// infinity. Use a pseudo-count philosophy.
// We have N sequences for distribution q. We say the frequency is less
// than 1/ N. We say the frequency is 1 / (N + 1).
func kl(kl_in *klIn, kl []float32) {
	logbase := math.Log(float64(kl_in.logbase))
	logb := func(x float64) float32 { // return logarithm base logbase
		return float32(math.Log(float64(x)) / logbase)
	}

	one_num_q := 1. / float64(kl_in.num_q+1)
	seqLen := len(kl_in.counts_p[0])
	nRow := len(kl_in.counts_p)
	for icol := 0; icol < seqLen; icol++ { //            icol position in seq
		for irow := 0; irow < nRow; irow++ { // irow is symbol
			pcount := float64(kl_in.counts_p[irow][icol])
			if pcount == 0 {
				continue
			}
			qcount := float64(kl_in.counts_q[irow][icol])
			if qcount == 0 {
				qcount = one_num_q // epsilon/pseudo-counts
			}
			tmp := logb(pcount / qcount)
			tmp *= float32(pcount)
			kl[icol] += tmp
		}
	}

}

// innerCosSim does the cosine product of two vectors
func innerCosSim(p, q []float32) float32 {
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

// calcCosSim calculates the cosine product of two distributions
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

	nRow := len(counts_p)

	for i := range counts_p[0] { //     i is position in sequence
		for j := 0; j < nRow; j++ { //  j is sequence number
			pvec[j] = counts_p[j][i]
			qvec[j] = counts_q[j][i]
		}
		r := innerCosSim(pvec, qvec)
		cosSim[i] = r
	}
}

// klFromSeqX takes a pair of seqXs and returns the KL
// distance in one direction. To get the other direction, call with
// the arguments reversed.
func klFromSeqX(seqXP, seqXQ *SeqX, klP []float32, wg *sync.WaitGroup) {
	defer wg.Done()
	klIn := klIn{
		counts_p: seqXP.counts.Mat,
		counts_q: seqXQ.counts.Mat,
		num_q:    seqXQ.nseq,
		logbase:  seqXP.logbase,
	}
	kl(&klIn, klP)
}

// sane checks some properties to catch errors
func sane(seqXP, seqXQ *SeqX, fileP, fileQ string) error {
	const mismatch = "Sequence length mismatch. %s: len %d, %s: len %d"
	const toofew = "File %s seems to have too few sequences"
	if seqXP.nseq <= 1 {
		return fmt.Errorf(toofew, fileP)
	}
	if seqXQ.nseq <= 1 {
		return fmt.Errorf(toofew, fileQ)
	}
	lenAln := seqXP.GetLen() // Check alignments have same length.
	if l2 := seqXQ.GetLen(); l2 != lenAln {
		return fmt.Errorf(mismatch, fileP, lenAln, fileQ, l2)
	}
	return nil
}

// entropywrap is a wrapper around the call to entropy, just
// to allow us to use a waitgroup
func entropyWrap(gapsAreChar bool, matrix [][]float32, entropy []float32,
	logbase int, gapMapping uint8, wg *sync.WaitGroup) {
	defer wg.Done()
	seq.EntropyFromArray(gapsAreChar, matrix, entropy, logbase, gapMapping)
}

// calcInner is the inner function to calculate entropies and KL distances.
// It could have been in main, but is in its own function so we can expose it
// for testing. We could avoid the five calls to make(). Call it once and divide
// into five pieces.
func calcInner(seqXP, seqXQ SeqX) (klP, klQ, entropyP, entropyQ, cosSim []float32) {
	var wg sync.WaitGroup

	klP = make([]float32, seqXP.GetLen())
	wg.Add(1)
	go klFromSeqX(&seqXP, &seqXQ, klP, &wg)

	klQ = make([]float32, seqXP.GetLen())
	wg.Add(1)
	go klFromSeqX(&seqXQ, &seqXP, klQ, &wg)

	const gapsAreChar = false
	entropyP = make([]float32, seqXP.GetLen())
	wg.Add(1) // race on next line
	go entropyWrap(gapsAreChar, seqXP.counts.Mat, entropyP, seqXP.logbase, seqXP.gapMapping, &wg)

	entropyQ = make([]float32, seqXP.GetLen())
	wg.Add(1)
	go entropyWrap(gapsAreChar, seqXQ.counts.Mat, entropyQ, seqXQ.logbase, seqXQ.gapMapping, &wg)

	cosSim = make([]float32, seqXP.GetLen())
	calcCosSim(seqXP.counts.Mat, seqXQ.counts.Mat, cosSim)

	wg.Wait()
	return klP, klQ, entropyP, entropyQ, cosSim
}

type numResults struct {
	klP, klQ, entropyP, entropyQ, cosSim []float32
}

// writeKl writes the results to a file
func writeKl(wrtr io.WriteCloser, numresults *numResults, offset int) {
	heading := `"res num", "klP", "klQ", "S_p", "S_q", "cosine sim"`
	fmt.Fprintln(wrtr, heading)
	for i := range numresults.klP {
		fmt.Fprintf(wrtr, "%d,%g,%g,%g,%g,%g\n", i+1+offset,
			numresults.klP[i], numresults.klQ[i],
			numresults.entropyP[i], numresults.entropyQ[i], numresults.cosSim[i])
	}
}

// Mymain is the main function for kullback-leibler distance
func Mymain(flags *CmdFlag, fileP, fileQ, outfile string) (err error) {
	var seqXP, seqXQ SeqX
	var wrtr io.WriteCloser
	if err := readtwofiles(flags, fileP, fileQ, &seqXP, &seqXQ); err != nil {
		return err
	}
	if err := sane(&seqXP, &seqXQ, fileP, fileQ); err != nil {
		return err
	}

	klP, klQ, entropyP, entropyQ, cosSim := calcInner(seqXP, seqXQ)
	if outfile == "" || outfile == "-" {
		wrtr = os.Stdout
	} else {
		fp, err := os.Create(outfile)
		if err != nil {
			return err
		}
		defer fp.Close()
		wrtr = fp
	}
	numresults := &numResults{
		klP: klP, klQ: klQ, entropyP: entropyP, entropyQ: entropyQ, cosSim: cosSim,
	}
	writeKl(wrtr, numresults, flags.Offset)

	return nil
}
