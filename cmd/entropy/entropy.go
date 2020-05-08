// 26 April 2020
// Read up a multiple sequence alignment and calculate the entropy
// per column.

package main

import (
	"flag"
	"fmt"
	. "github.com/andrew-torda/goutil/seq/common"
	"github.com/andrew-torda/seq_compat/pkg/entropy"
	"os"
	"path"
)

func usage() int {
	fmt.Fprintln(os.Stderr, "usage:", path.Base(os.Args[0]), "[infile [outfile]]")
	flag.PrintDefaults()
	return (ExitUsageError)
}

func main() {
	var flags entropy.CmdFlag
	var infile, outfile string
	flag.IntVar(&flags.Offset, "f", 0, "offset for numbering output")
	flag.BoolVar(&flags.GapsAreChar, "g", false, "gap is a valid symbol")
	flag.IntVar(&flags.NSym, "n", -1, "num symbols, guessed by default, 4 for DNA")
	flag.StringVar(&outfile, "o", "", "output file name, default stdout")
	flag.StringVar(&flags.RefSeq, "r", "", "reference sequence, check compatibility")

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
