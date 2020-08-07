// 29 April 2020

package squash

import (
	"fmt"
	"github.com/andrew-torda/seq_compat/pkg/seq"
	. "github.com/andrew-torda/seq_compat/pkg/seq/common"
	"os"
)

// MyMain is the top level main, after parsing the command line.
func MyMain(seqstring, infile, outfile string) int {
	s_opts := &seq.Options{
		Vbsty: 0, Keep_gaps_rd: true,
		Dry_run:      false,
		Rmv_gaps_wrt: false}

	seqgrp, err := seq.Readfile(infile, s_opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return ExitFailure
	}
	var ndxref int
	if ndxref = seqgrp.FindNdx(seqstring); ndxref == -1 {
		fmt.Fprintf(os.Stderr, "Could not find \"%s\" amongst sequences\n", seqstring)
		return ExitFailure
	}
	seqslc := seqgrp.GetSeqSlc()       // Slice of sequences
	maskseq := seqslc[ndxref].GetSeq() // The reference sequence
	nfullseq := len(maskseq)
	mask := make([]bool, nfullseq)
	for i := range mask { // First, we assume all sites are interesting
		mask[i] = true
	}
	for i, c := range maskseq { // Then we pick out the few that
		if c == GapChar { //       should be taken out.
			mask[i] = false
		}
	}
	const emsg = "Length mismatch ref: %d seq %d len %d\n"
	for i, ss := range seqgrp.GetSeqSlc() {
		if len(ss.GetSeq()) != nfullseq {
			fmt.Fprintf(os.Stderr, emsg, nfullseq, i, len(ss.GetSeq()))
			return ExitFailure
		}
		b := ss.GetSeq()[:0]
		for i, c := range ss.GetSeq() {
			if mask[i] {
				b = append(b, c)
			}
		}
		seqgrp.GetSeqSlc()[i].SetSeq(b) // do not use "ss" here
	}

	if err := seq.WriteToF(outfile, seqgrp.GetSeqSlc(), s_opts); err != nil {
		if outfile == "" {
			outfile = "os.Stdout"
		}
		fmt.Fprintln(os.Stderr, "Fail writing to ", outfile, err)
		return ExitFailure
	}

	return ExitSuccess
}
