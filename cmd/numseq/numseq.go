// 3 Aug 2020

// Open a file and count the number of ">" characters. This might be
// The number of sequences.

package main

import (
	"fmt"
	"log"
	"os"
	"github.com/andrew-torda/seq_compat/pkg/numseq"
	. "github.com/andrew-torda/seq_compat/pkg/seq/common"
)

func usage () {
	fmt.Fprintln (os.Stderr, "usage:", os.Args[0], "filename")
}

func main() {
	if len (os.Args) < 2 {
		usage()
		os.Exit (ExitUsageError)
	}
	fname := os.Args[1]
	var err error
	var nOccur int
	if nOccur, err = numseq.Main(fname); err != nil {
		log.Fatal (err)
	}
	
	fmt.Println ("got", nOccur)
	os.Exit (ExitSuccess)
}
