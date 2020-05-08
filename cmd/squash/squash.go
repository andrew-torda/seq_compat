// 29 April 2020
// Squash an alignment using some specified sequence as a reference.

package main

import (
	"fmt"
	"os"
	"path"

	. "github.com/andrew-torda/goutil/seq/common"
	"github.com/andrew-torda/seq_compat/pkg/squash"
)

// usage
func usage() {
	name := path.Base(os.Args[0])
	fmt.Println("usage:", name, "sequence_string [inputfile [outputfile]]")
	os.Exit(ExitFailure)
}

func main() {
	var seqstring, infile, outfile string
	nArg := len(os.Args)
	if nArg <= 1 {
		fmt.Fprintln(os.Stderr, "require at least one command line argument")
		usage()
	}
	seqstring = os.Args[1]
	if nArg > 2 {
		infile = os.Args[2]
	}
	if nArg > 3 {
		outfile = os.Args[3]
	}
	os.Exit(squash.MyMain(seqstring, infile, outfile))
}
