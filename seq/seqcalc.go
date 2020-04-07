// 6 Apr 2020
// seqcalc does simple, common calculations on a set of sequences.
// The functions have to live in this package, since they
// need access to the internals of a sequence

package seq

// SetSymUsed fills out the bool slice which says whether or not a
// symbol was used
func (seqgrp *SeqGrp) SetSymUsed() {
	for _, ss := range seqgrp.seqs {
		s := ss.GetSeq()
		for _, c := range s {
			seqgrp.symUsed[c] = true
		}
	}
	seqgrp.usedKnwn = true
}


// GetType looks at a set of sequences and returns its best guess
// as to the type of file.
func (seqgrp *SeqGrp) GetType() SeqType {
	if seqgrp.usedKnwn != true {
		seqgrp.SetSymUsed()
	}
	if seqgrp.stype != Unknown { // If the sequence type has been
		return seqgrp.stype //      set, just return it.
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

// GetCount is given an array of MaxSym size (probably 127). It returns
// an array with the number of times each symbol is used at each position.
func (seqgrp *SeqGrp) getCount() error {
	return nil
}
