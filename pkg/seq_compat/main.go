// 4 April 2020
package seq_compat

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/andrew-torda/goutil/seq"
	"github.com/andrew-torda/seq_compat/pkg/seqcalc"
)

const (
	ExitSuccess = iota
	ExitFailure
	ExitUsageError
)

func usage() int {
	fmt.Fprintln(os.Stderr, "usage:", path.Base(os.Args[0]), "infile plotfile")
	flag.PrintDefaults()
	return (ExitUsageError)
}

func Mymain() int {
	noDupCheck := flag.Bool("d", false, "do not check for duplicate sequences")
	gapFlag := flag.Bool("i", false, "ignore blanks")
	refSeqNdx := flag.Int("n", 1, "index of reference sequence (index from 1)")
	vbsty := flag.Int("v", 3, "verbosity")

	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Println("Got", len(flag.Args()), "args, expected 2")
		return (usage())
	}
	infile := flag.Args()[0]
	outfile := flag.Args()[1]

	fmt.Println("gapflag is", *gapFlag, "in", infile, "outfile", outfile)
	if *noDupCheck {
		fmt.Println("Will check for duplicate sequences")
	} else {
		fmt.Println("Not checking for duplicate sequences")
	}
	fmt.Println("reference sequence is", *refSeqNdx)

	seq_map := make(map[string]int)
	seq_set := make([]seq.Seq, 0, 0)
	var n_dup int
	s_opts := &seq.Options{Vbsty: *vbsty, Keep_gaps: true}

	seqs, n_dup, err := seq.ReadSeqs(infile, seq_set[:], seq_map, s_opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: %v", err)
		return (ExitFailure)
	}
	if n_dup > 0 {
		fmt.Fprintln(os.Stderr, "Input probably has duplicate sequences")
	}
	seqcalc.Calc(seqs, 1)
	fmt.Println(seqs)

	return (ExitSuccess)
}
