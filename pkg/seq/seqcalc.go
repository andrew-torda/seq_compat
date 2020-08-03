// 6 Apr 2020
// seqcalc does simple, common calculations on a set of sequences.
// The functions have to live in this package, since they
// need access to the internals of a sequence

package seq

import (
	"math"
	"sync"

	. "github.com/andrew-torda/seq_compat/pkg/seq/common"
	"github.com/andrew-torda/matrix"
)

const (
	badMap = math.MaxUint8 // marks a symbol as not seen
)

type SymSync struct {
	Once  sync.Once
	UChan chan [MaxSym]bool
}

// mergelists merges two lists of symbols that have been
// used. It reads each list from a channel, merges them
// and returns the merged list, which will have overwritten
// the first list it received.
func mergelists(uChan chan [MaxSym]bool) {
	a1, a2 := <-uChan, <-uChan
	for i := range a1 {
		a1[i] = a1[i] || a2[i]
	}
	uChan <- a1
	uChan <- a1
	close(uChan)
}

// SetSymUsed fills out the bool slice which says whether or not a
// symbol was used.
// Normally, this is just a loop over all sequences. If we are combining
// two seqgrp's, then the symbols used in group A should also be
// marked used in group B and vice versa. If we get a second varadic
// argument, it is a channel to be used in combining.
func (seqgrp *SeqGrp) SetSymUsed(symSync ...*SymSync) {
	for _, ss := range seqgrp.seqs {
		s := ss.GetSeq()
		for _, c := range s {
			seqgrp.symUsed[c] = true
		}
	}
	if symSync != nil {
		go symSync[0].Once.Do(func() { mergelists(symSync[0].UChan) })
		symSync[0].UChan <- seqgrp.symUsed
		seqgrp.symUsed = <-symSync[0].UChan
	}
	seqgrp.usedKnwn = true
}

// GetType looks at a set of sequences and returns its best guess
// as to the type of file.
func (seqgrp *SeqGrp) GetType() SeqType {
	if seqgrp.stype != Unchecked { // If the sequence type has been
		return seqgrp.stype //      set, just return it.
	}

	if seqgrp.usedKnwn != true {
		seqgrp.SetSymUsed()
	}
	protType := []byte{
		'D', 'E', 'F', 'H', 'I', 'K', 'L', 'M',
		'N', 'P', 'Q', 'R', 'S', 'V', 'W', 'Y'}

	used := seqgrp.symUsed
	for _, c := range protType { // If we see an amino acid code,
		if used[c] { //          just return protein type.
			return Protein
		}
	}

	if used['T'] && used['U'] {
		return Ntide
	}
	// If we have ACG, but neither T or U, it is a nucleotide
	// but we cannot tell if it is RNA or DNA
	if used['A'] && used['C'] && used['G'] && !used['T'] && !used['U'] {
		return Ntide
	}
	if used['T'] {
		return DNA
	}
	if used['U'] {
		return RNA
	}

	return Unknown
}

// mapsyms looks at the symbols(characters, bases, residues) used in a
// seqgrp. It then makes a little array for mapping.
func (seqgrp *SeqGrp) mapsyms() error {
	if seqgrp.usedKnwn != true {
		seqgrp.SetSymUsed()
	}
	for i := range seqgrp.mapping { // Initialise with bad value, to
		seqgrp.mapping[i] = badMap // trap errors later
	}

	var n uint8
	for i := range seqgrp.symUsed {
		if seqgrp.symUsed[i] {
			seqgrp.mapping[i] = n
			if n >= badMap {
				panic("program bug")
			}
			seqgrp.revmap = append(seqgrp.revmap, uint8(i))
			n++
		}
	}
	return nil
}

// UsageSite counts how many of each symbol/character appear at
// each site in the alignment.
// counts.Mat looks like [length_of_seq][number_of_types]
// We store it as a float32, since it will later usually be normalised
// and converted to a fraction.
// Inaccuracy introduced by working with floats is no problem and we
// can avoid allocating a new matrix for the frequencies.
func (seqgrp *SeqGrp) UsageSite() {
	if len(seqgrp.revmap) == 0 {
		seqgrp.mapsyms()
	}
	nrow := len(seqgrp.revmap)
	ncol := len(seqgrp.seqs[0].GetSeq())
	seqgrp.counts = matrix.NewFMatrix2d(nrow, ncol)
	for _, ss := range seqgrp.seqs {
		for i, c := range ss.GetSeq() {
			cmap := seqgrp.mapping[c]
			seqgrp.counts.Mat[cmap][i] += 1
		}
	}
}

// Usage Frac converts count to normalised frequencies. If letter 'A'
// occurs 2 times in five positions, its count entry will be changed from
// 2 to 2/5 = 0.4
// If gapsAreChar is true, gaps ("-") are treated as a valid character
// type. Otherwise they are removed from the tallies.
// If gapsAreChar is not true, then
//    a symbol's fraction is the fraction of non-gaps
//                in which you find this symbol
//    the gap's fraction is the fraction of the total
//                number of residues in which one finds a gap.
// This means that the fractions of non-gaps adds up to 1,
// and then you have a bit more due to gaps.
// It also means that the data looks correct when you plot it out.
func (seqgrp *SeqGrp) UsageFrac(gapsAreChar bool) {
	if seqgrp.counts == nil {
		seqgrp.UsageSite()
	}
	counts := seqgrp.counts
	gappos := seqgrp.mapping[GapChar]
	thereAreGaps := true
	if gappos == badMap {
		thereAreGaps = false
	}
	nrow, ncol := counts.Size()
	total := make([]float32, ncol) // total observations in each column
	for icol := 0; icol < ncol; icol++ {
		for irow := 0; irow < nrow; irow++ {
			total[icol] += counts.Mat[irow][icol]
		}
	}
	var savedGapFrac []float32
	if thereAreGaps {
		if gapsAreChar == false {
			savedGapFrac = make([]float32, ncol)
			for icol := range savedGapFrac {
				savedGapFrac[icol] = counts.Mat[gappos][icol] / total[icol]
			}
			for icol := 0; icol < ncol; icol++ { // Remove gaps from the totals
				total[icol] -= counts.Mat[gappos][icol]
			}
		}
	}
	for icol := 0; icol < ncol; icol++ { // Normalise the counts
		for irow := 0; irow < nrow; irow++ {
			if total[icol] != 0 {
				counts.Mat[irow][icol] /= (total[icol])
			}
		}
	}
	// The gaps have to be corrected. They have to be a fraction of the
	// original column totals
	for icol := range savedGapFrac {
		counts.Mat[gappos][icol] = savedGapFrac[icol]
	}
	seqgrp.freqKnwn = true
}

// GapFrac looks in a SeqGrp and returns a slice with the fraction
// of gap characters at each position. If there are no gaps, there
// is no slice so we quietly return nil without signalling an error.
func (seqgrp *SeqGrp) GapFrac() []float32 {
	if !seqgrp.freqKnwn {
		gapsAreChar := true // Does not matter what we say here
		seqgrp.UsageFrac(gapsAreChar)
	}
	gappos := seqgrp.mapping[GapChar]
	if gappos == badMap {
		return nil
	}
	return seqgrp.counts.Mat[gappos]
}

// GetLogBase returns the base to be used for logarithms
func (seqgrp *SeqGrp) GetLogBase(gapsAreChar bool) (nSym int) {
	const progbug = "program bug in GetLogBase"
	if !seqgrp.usedKnwn {
		seqgrp.UsageSite()
	}
	if gapsAreChar {
		switch seqgrp.GetType() {
		case DNA, RNA, Ntide:
			nSym = 5 // 4 nucleotides + gap character
		case Protein:
			nSym = 21 // 20 residues, + gap
		case Unknown:
			nSym = len(seqgrp.revmap)
		default:
			panic(progbug)
		}
	} else { // gaps are ignored
		switch seqgrp.GetType() {
		case DNA, RNA, Ntide:
			nSym = 4
		case Protein:
			nSym = 20
		case Unknown:
			nSym = len(seqgrp.revmap)
		default:
			panic(progbug)
		}
	}
	return nSym
}

// EntropyFromArray is the inner routine for calculating entropy.
// It operates on the inner matrix, so it can be called from other routines
// which do not have the seqgrp, but do have a table of counts.
func EntropyFromArray(gapsAreChar bool,
	matrix [][]float32, entropy []float32, logbase int, gapMapping uint8) {
	logfac := 1.0 / math.Log(float64(logbase)) // to change base of logs
	nrow := len(matrix)
	ncol := len(matrix[0])

	//	nrow, ncol := seqgrp.counts.Size()

	if gapsAreChar { //                         Gaps are just a character, so
		for icol := 0; icol < ncol; icol++ { // no need for special treatment
			total := 0.0
			for irow := 0; irow < nrow; irow++ {
				f := float64(matrix[irow][icol])
				if f == 0.0 {
					continue
				}
				logf := math.Log(f) * logfac
				total += f * logf
			}
			entropy[icol] = float32(math.Abs(total))
		}
	} else { // we have to check and ignore gap characters
		iBadRow := int(gapMapping)
		for icol := 0; icol < ncol; icol++ {
			total := 0.0
			for irow := 0; irow < nrow; irow++ {
				if irow == iBadRow {
					continue
				}
				f := float64(matrix[irow][icol])
				if f == 0.0 {
					continue
				}
				logf := math.Log(f) * logfac
				total += f * logf
			}
			entropy[icol] = -float32(total)
		}

	}
}

// Entropy calculates sequence entropy. It returns the result as a slice
// of the same length as the sequences. It needs to be told if gaps are
// characters, or should be ignored.
// If the sequence is a nucleotide or protein, we know what logarithm to use.
// If the sequence is unknown, we use the log base the number different
// symbols
// The caller allocates space for the result (entropy).
func (seqgrp *SeqGrp) Entropy(gapsAreChar bool, entropy []float32) {
	if !seqgrp.freqKnwn {
		seqgrp.UsageFrac(gapsAreChar)
	}
	logbase := seqgrp.GetLogBase(gapsAreChar)
	gapMapping := seqgrp.GetMapping(GapChar)
	EntropyFromArray(gapsAreChar, seqgrp.counts.Mat, entropy, logbase, gapMapping)
}

// Compat takes one sequence (a reference). It returns the frequency of each
// character from this sequence at each position in the alignment.
// Do you want to remove the reference sequence from the calculations ?
// Usually yes.
func (seqgrp *SeqGrp) Compat(refseq []byte, gapsAreChar bool) []float32 {
	if !seqgrp.freqKnwn { // Make sure symbol frequencies have been calculated
		seqgrp.UsageFrac(gapsAreChar)
	}
	compat := make([]float32, len(seqgrp.seqs[0].GetSeq()))
	ntotal := seqgrp.GetNSeq()
	gapfrac := seqgrp.GapFrac()
	if gapfrac == nil {
		gapfrac = make([]float32, len(seqgrp.seqs[0].GetSeq()))
	}

	for i, c := range refseq {
		if c == GapChar {
			compat[i] = 0
			continue
		}
		nseq := (1 - gapfrac[i]) * float32(ntotal)
		if nseq < 1.001 { // It means, we have a lonely insertion
			compat[i] = 0
		} else {
			ic := seqgrp.GetMap(c)
			fracC := seqgrp.counts.Mat[ic][i]
			nthischar := fracC*nseq - 1
			compat[i] = nthischar / (nseq - 1)
		}
	}
	return compat
}
