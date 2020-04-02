// 1 Apr 2019
// Read a sequence alignment. Take an arbitrary sequence as the one to
// be checked for compatibility using the Kuhlbach-Leibler divergence.

package main

import (
	"flag"
	"fmt"
	"github.com/andrew-torda/goutil/seq"
	"os"
	"path"
)

const (
	exitSuccess = 0
	exitFailure = iota
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage:", path.Base(os.Args[0]), "infile plotfile")
	flag.PrintDefaults()
}

func mymain() int {
	noDupCheck := flag.Bool("d", false, "do not check for duplicate sequences")
	gapFlag := flag.Bool("i", false, "ignore blanks")
	refseqID := flag.Int("n", 1, "index of reference sequence (index from 1)")
	fmt.Println("argv is ", os.Args)
	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Println("Got", len(flag.Args()), "args, expected 2")
		usage()
		return (exitFailure)
	}
	infile := flag.Args()[0]
	outfile := flag.Args()[1]

	fmt.Println("gapflag is", *gapFlag, "in", infile, "outfile", outfile)
	if *noDupCheck {
		fmt.Println("Will check for duplicate sequences")
	} else {
		fmt.Println("Not checking for duplicate sequences")
	}
	fmt.Println("reference sequence is", *refseqID)
	seq_map := make(map[string]int)
	seq_set := make([]seq.Seq, 0, 0)
	var n_dup int

	s_opts := &seq.Options{Vbsty: 3, Keep_gaps: true}

	seqs, n_dup, err := seq.ReadSeqs(infile, seq_set[:], seq_map, s_opts)
	if err != nil {
		fmt.Fprintln (os.Stderr, "Fatal: %v", err)
		return (exitFailure)
	}
	if n_dup > 0 {
		fmt.Fprintln (os.Stderr, "Input probably has duplicate sequences")
	}
	fmt.Println (seqs)
	return (exitSuccess)
}

func main() {
	os.Exit(mymain())
}
