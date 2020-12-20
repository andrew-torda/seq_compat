// 26 April 2020
// Read up a multiple sequence alignment and calculate the entropy
// per column.

package main

import (
	"flag"
	"fmt"
	"github.com/andrew-torda/seq_compat/pkg/entropy"
	. "github.com/andrew-torda/seq_compat/pkg/seq/common"
	"os"
	"path"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage:", path.Base(os.Args[0]), "[infile [outfile]]")
	long := `Do not just type type command name. It will wait on input from stdin.
Given no arguments, read and write from stdin / stdout.
Given one argument, read from the given file name, but write to stdout.
Given two arguments, read from the first one, write to the second.`
	fmt.Fprintln(os.Stderr, long)
	flag.PrintDefaults()
}

func main() {
	var flags entropy.CmdFlag
	var infile, outfile string

	flag.StringVar(&flags.Chimera, "c", "", "filename to write chimera format to")
	flag.IntVar(&flags.Offset, "f", 0, "offset for numbering output, renumbering sites")
	flag.BoolVar(&flags.GapsAreChar, "g", false, "gap is a valid symbol")
	flag.IntVar(&flags.NSym, "n", -1, "num symbols, guessed by default, 4 for DNA")
	flag.StringVar(&flags.RefSeq, "r", "", "reference sequence, check compatibility")
	flag.BoolVar(&flags.Time, "t", false, "print out timing information")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() > 0 {
		infile = flag.Arg(0)
		if flag.NArg() > 1 {
			outfile = flag.Arg(1)
		}
	}

	if err := entropy.Mymain(&flags, infile, outfile); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitFailure)
	} else {
		os.Exit(ExitSuccess)
	}
}
