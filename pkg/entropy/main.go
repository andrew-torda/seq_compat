// 27 april 2020
package entropy

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/andrew-torda/seq_compat/pkg/seq"
)

type ntrpyargs struct {
	entropy []float32 // sequence entropy
	gapfrac []float32 // fraction of gap entries in column
	compat  []float32 // compatibility of reference sequence
	outfile string    // write to a file or standard input // Check.. is this used ? XX
	refseq  []byte    // nil or reference sequence
	offset  int       // residue number offset on output
}

// writeNtrpy write a simple entropy file.
func writeNtrpy(args *ntrpyargs) error {
	headings1 := `"res num","entropy","%frac non-gap"`
	if args.refseq != nil {
		headings1 += `,"res name","compatibility"`
	}
	if _, err := os.Stat(args.outfile); err == nil {
		fmt.Fprintln(os.Stderr, "Warning, trashing old version of", args.outfile)
	}
	var fp io.WriteCloser
	var err error
	if args.outfile != "" {
		if fp, err = os.Create(args.outfile); err != nil {
			return fmt.Errorf("output file %v: %w", args.outfile, err)
		}
		defer fp.Close()
	} else {
		fp = os.Stdout
	}
	// set up the gaps.
	if args.gapfrac == nil { // Could be that there are no gaps.
		args.gapfrac = make([]float32, len(args.entropy))
		for i, _ := range args.entropy { // To avoid if statements below
			args.gapfrac[i] = 0 // make an array and just fill it with zeroes.
		}
	}
	fmt.Fprintln(fp, headings1)
	for i, v := range args.entropy {
		fmt.Fprintf(fp, "%d,%.2f,%.2f", i+1+args.offset, v, 1-args.gapfrac[i])
		if args.refseq != nil {
			fmt.Fprintf(fp, ",%c,%.2f", args.refseq[i], args.compat[i])
		}
		fmt.Fprintln(fp)
	}
	return nil
}

type CmdFlag struct {
	Offset      int    // Add this to the residue numbering on output
	GapsAreChar bool   // Do we keep gaps ? Are gaps a valid symbol ?
	NSym        int    // Set the number of symbols in sequences
	RefSeq      string // A reference seq, whose compatibility will be calculated
	Time        bool   // do we want to print out run time ?
}

// Mymain is the main function for calculating entropy and writing to a file
func Mymain(flags *CmdFlag, infile, outfile string) error {
	var err error
	s_opts := &seq.Options{}
	if flags.Time {
		startTime := time.Now()
		end := func () { // Wrapping in a closure seems to be helpful. Gives the right time.
			fmt.Println("finished after", time.Since(startTime).Milliseconds(), "ms")
		}
		defer end()
	}
	var ntrpyargs = &ntrpyargs{ // start setting up things to go
		outfile: outfile, // to the printing function later
		offset:  flags.Offset}

	seqgrp, err := seq.Readfile(infile, s_opts)
	if err != nil {
		return (fmt.Errorf("Fail reading sequences: %w", err))
	}

	if flags.RefSeq != "" {
		if ndxSeq := seqgrp.FindNdx(flags.RefSeq); ndxSeq == -1 {
			return (fmt.Errorf(`Cannot find ref sequence "%s"\n`, flags.RefSeq))
		} else {
			ntrpyargs.refseq = seqgrp.SeqSlc()[ndxSeq].GetSeq()
			ntrpyargs.compat = seqgrp.Compat(ntrpyargs.refseq, flags.GapsAreChar)
		}
	}
	seqgrp.Upper()
	ntrpyargs.gapfrac = seqgrp.GapFrac()
	ntrpyargs.entropy = make([]float32, seqgrp.GetLen())
	seqgrp.Entropy(flags.GapsAreChar, ntrpyargs.entropy)

	if err = writeNtrpy(ntrpyargs); err != nil {
		return err
	}
	return nil
}
