// The main part of the program for sequence compatibility

package seqcalc

import (
	"fmt"
	"unsafe"

	"github.com/andrew-torda/goutil/seq"
)

// function Calc
func Calc(seqs []seq.Seq, refseqNdx int) error {
	counts := make([]int, len(seqs[0].GetSeq()))
	for snum, s := range seqs { // Make all sequences upper case
		if err := s.Upper(); err != nil {
			err := fmt.Errorf("Sequence number %d: %w", snum, err)
			return err
		}
	}
	refseqNdx = refseqNdx - 1
	//	refseq := seqs[refseqNdx]           // save reference sequence
	seqs[refseqNdx] = seqs[len(seqs)-1] // move last element into his place
	seqs = seqs[:len(seqs)-1]           // shorten by one position

	fmt.Println(seqs, counts, seq.Unknown)
	return nil
}
