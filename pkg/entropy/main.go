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
	outfile string    // write to a file or standard input // Check.. is this used ? X
	refseq  []byte    // nil or reference sequence
	offset  int       // residue number offset on output
}

// warnExists checks if a filename exists and prints a warning
// if we will trash a file. It does not return an error.
func warnExists(fname string) {
	if _, err := os.Stat(fname); err == nil {
		fmt.Fprintln(os.Stderr, "Warning, trashing old version of", fname)
	}
}

// wrtAtt writes an array of numbers to an open file pointer in the format
// wanted by chimera for attributes.
func wrtAtt(fp io.WriteCloser, attname string, tmpNums []float32, offset int) error {
	head := "\nattribute: " + attname + "\nmatch mode: 1-to-1\nrecipient: residues"
	fmt.Fprintln(fp, "#", time.Now().Format(time.RFC1123), head)
	for i, v := range tmpNums {
		rnum := i + 1 + offset
		if _, err := fmt.Fprintf(fp, "\t:%d\t%#g\n", rnum, v); err != nil {
			return err
		}
	}
	return nil
}

// interesting is a hack, but useful. If a residue is present more than
// 60 % of the time, save its entropy. If it is not present so often,
// set interesting value to 0.5
func interesting(args *ntrpyargs, tmpnum []float32) {
	for i, entropy := range args.entropy {
		if args.gapfrac[i] > 0.4 {
			tmpnum[i] = entropy
		} else {
			tmpnum[i] = 0.5
		}
	}
}

// writeChimera writes the entropy information in a form suitable
// for reading in chimera as an attribute file
func writeChimera(fname string, args *ntrpyargs) error {
	var fp io.WriteCloser
	var err error

	if fname == "-" { // Write to stdout
		fp = os.Stdout
	} else { //            Write to a named file
		warnExists(fname)
		if fp, err = os.Create(fname); err != nil {
			return fmt.Errorf("chimera output file %v: %w", fname, err)
		}
		defer fp.Close()
	}
	if err = wrtAtt(fp, "entropy", args.entropy, args.offset); err != nil {
		return err
	}
	tmpnum := make([]float32, len(args.gapfrac))
	for i, v := range args.gapfrac { // For plotting, we probably do not
		tmpnum[i] = 1 - v //            gaps, but rather 1 - fraction of gaps
	}
	if err = wrtAtt(fp, "present", tmpnum, args.offset); err != nil {
		return err
	}
	interesting(args, tmpnum)
	if err = wrtAtt(fp, "interesting", tmpnum, args.offset); err != nil {
		return err
	}
	return nil
}

// writeNtrpy write a simple entropy file. If there is no filename or the
// filename is "-", write to standard output.
func writeNtrpy(args *ntrpyargs) error {
	headings1 := `"res num","entropy","%frac non-gap"`
	if args.refseq != nil {
		headings1 += `,"res name","compatibility"`
	}
	warnExists(args.outfile)
	var fp io.WriteCloser
	var err error
	if args.outfile != "" && args.outfile != "-" {
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
	Chimera     string // write output in format for chimera
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
		end := func() { // Wrapping in a closure is helpful. Gives the right time.
			fmt.Println("finished after", time.Since(startTime).Milliseconds(), "ms")
		}
		defer end()
	}
	var ntrpyargs = &ntrpyargs{ // start setting up things to go
		outfile: outfile, //       to the printing function later
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
	if flags.Chimera != "" { // Do we have to write a chimera attribute file ?
		if err = writeChimera(flags.Chimera, ntrpyargs); err != nil {
			return err
		}
	}
	return nil
}
