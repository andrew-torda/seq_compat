// 25 may 2025
// seqlen visits a fasta file and counts the length of each sequence
// after removing gaps.

package main

import (
	"os"
	"fmt"
	"flag"

	"github.com/andrew-torda/seq_compat/pkg/seqlen"
)

const (
	ExitSuccess = 0
	ExitFailure = 1
)

func main () {
	uStr := "usage: seqlen [options] input output"
	var cmdArgs seqlen.CmdArgs
	flag.BoolVar (&cmdArgs.IgnrSeqLen, "i", false, "ignore sequence lengths not being consistent")
	flag.StringVar(&cmdArgs.OutSeqFname, "s", "", "Write cleaned sequences to")
	flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [options] input output\n\n", os.Args[0])
        flag.PrintDefaults()
    }

	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintln (os.Stderr, "Expected two arguments. Got ", flag.NArg())
		fmt.Fprintln (os.Stderr, uStr)
	}
	cmdArgs.InSeqFname = flag.Arg(0)
	cmdArgs.OutCntFname = flag.Arg(1)
	if err := seqlen.Mymain(cmdArgs); err != nil {
		fmt.Fprintln (os.Stderr, err)
		os.Exit (ExitFailure)
	}
	os.Exit(ExitSuccess)
}
