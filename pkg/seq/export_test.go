package seq

import (
	"fmt"
)

// PrintFreqs prints out the frequencies of each character type
// It is probably only useful for debugging or testing.
// format is a format string like "%6.1f"
func (seqgrp *SeqGrp) PrintFreqs(format string) {
	if len(seqgrp.revmap) == 0 {
		seqgrp.UsageSite()
	}
	for ic, m := range seqgrp.revmap {
		fmt.Printf("%c ", m)
		for i := 0; i < len(seqgrp.seqs[0].seq); i++ {
			fmt.Printf(format, seqgrp.counts.Mat[ic][i])
		}
		fmt.Printf("\n")
	}
}

var SetFastaRdSize = setFastaRdSize

func (seqgrp *SeqGrp) Clear() { seqgrp.clear() }
