// 31 July 2020

package main

import (
	"flag"
	"fmt"
	"strconv"
	"os"

	"github.com/andrew-torda/seq_compat/pkg/randseq"
	. "github.com/andrew-torda/seq_compat/pkg/seq/common"
)

func main() {
	f := flag.NewFlagSet("randseq", flag.ExitOnError)
	const iseed int64 = 1637
	var args randseq.RandSeqArgs

	f.BoolVar(&args.NoGap, "g", false, "do not put gaps in sequences")
	f.BoolVar(&args.MkErr, "e", false, "provoke errors")
	f.Int64Var(&args.Iseed, "r", iseed, "random number seed")
	if err := f.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln (f.Output(), err)
		os.Exit(ExitUsageError)
	}
	if f.NArg() != 3 {
		fmt.Fprintln(f.Output(), "Too few args\nrandseq [..] file nseq length")
		f.Usage()
		
		os.Exit(ExitUsageError)
	}

	fname := f.Args()[0]
	if fname == "-" || fname == "" {
		args.Wrtr = os.Stdout
	} else {
		if ft, err := os.Create(f.Args()[0]); err != nil {
			fmt.Fprintln(os.Stderr, "File for output:", err)
			os.Exit(1)
		} else {
			defer ft.Close()
			args.Wrtr = ft
		}
	}

	const emsg = "Failed converting %s to positive integer"
	if nseq, err := strconv.ParseUint(f.Args()[1], 10, 32); err != nil {
		fmt.Fprintf(os.Stderr, emsg, f.Args()[1])
		os.Exit(ExitFailure)
	} else {
		args.Nseq = int(nseq)
	}
	if nlen, err := strconv.ParseUint(f.Args()[2], 10, 32); err != nil {
		fmt.Fprintf(os.Stderr, emsg, f.Args()[2])
		os.Exit(ExitFailure)
	} else {
		args.Len = int(nlen)
	}
	if err := randseq.RandSeqMain(&args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitFailure)
	} else {
		os.Exit(ExitSuccess)
	}
}
