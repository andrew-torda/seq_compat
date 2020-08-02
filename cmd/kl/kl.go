// 28 May 2020

// When I write this, it should
//   read up two alignments
//   calculate kullback-leibler divergence for p against q, then q against p
//   Write out a long list
//     kl_pq, kl_qp, entropy p, entropy q, dot-product, jensen-blah distance,
// If we code this carefully, we can re-use the existing entropy code.

package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	. "github.com/andrew-torda/goutil/seq/common"
	"github.com/andrew-torda/seq_compat/pkg/kl"
)

// usage
func usage() int {
	fmt.Fprintln(os.Stderr, "usage:", path.Base(os.Args[0]), "[opts] f1.msa f2.msa")
	flag.PrintDefaults()
	return (ExitUsageError)
}

// main
func main() {
	var flags kl.CmdFlag
	outfile := "-"
	flag.IntVar(&flags.Offset, "f", 0, "offset for numbering output")
	flag.BoolVar(&flags.GapsAreChar, "g", false, "gap is a valid symbol")
	flag.IntVar(&flags.NSym, "n", -1, "num symbols, guessed by default, 4 for DNA")
	flag.StringVar(&outfile, "o", "", "output file name, default stdout")

	flag.Parse()

	seqf1 := flag.Arg(0)
	seqf2 := flag.Arg(1)
	if seqf1 == "" || seqf2 == "" {
		os.Exit(usage())
	}
	if err := kl.Mymain(&flags, seqf1, seqf2, outfile); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitFailure)
	} else {
		os.Exit(ExitSuccess)
	}
}
