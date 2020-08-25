// Compare readfasta before and after doing the single block for storing
// sequences.
// Dam syntax... dlv test -- -test.bench ForMem
// go test -bench ForMem -memprofile mem.out
// From command line
package seq_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/andrew-torda/seq_compat/pkg/seq"
)

func writeTmpSeqFile() (*os.File, error) {
	fp, err := ioutil.TempFile(".", "del_me")
	if err != nil {
		return nil, err
	}
	nseq := 214651
	nrep := 27
	for i := 0; i < nseq; i++ {
		fmt.Fprintln(fp, "> seq", i)
		for j := 0; j < nrep; j++ {
			fmt.Fprint(fp, "aaaaaaaaaaaaa ")
		}
		fmt.Fprint(fp, "\n")
	}
	fp.Seek(0, 0)
	return fp, nil
}

func BenchmarkForMemUse(b *testing.B) {
	fp, err := writeTmpSeqFile()
	if err != nil {
		b.Fatal("program bug")
	}
	frem := func() { os.Remove(fp.Name()) }
	b.Cleanup(frem)
	var seqgrp seq.SeqGrp
	s_opts := &seq.Options{}

	if err = seq.ReadFasta(fp, &seqgrp, s_opts); err != nil {
		b.Fatal("benchmark broke reading sequences")
	}
	fp.Close()
}
