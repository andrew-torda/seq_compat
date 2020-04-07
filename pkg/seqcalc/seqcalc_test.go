// 3 Apr 2020

package seqcalc_test

import (
	"fmt"
	"testing"

	"github.com/andrew-torda/goutil/seq"
	. "github.com/andrew-torda/seq_compat/pkg/seqcalc"
)

func TestBadSym(t *testing.T) {
	infile := "testdata/badsym.fa"
	s_opts := &seq.Options{Vbsty: 3, Keep_gaps: true}
	seq_map := make(map[string]int)
	seq_set := make([]seq.Seq, 0, 0)

	seqs, _, err := seq.ReadSeqs(infile, seq_set[:], seq_map, s_opts)
	if err != nil {
		t.Errorf(err.Error())
	}
	const refSeqNdx = 1
	if err = Calc(seqs, refSeqNdx); err == nil {
		t.Errorf("Bad symbol not recognised")
	}
}

func TestSimple(t *testing.T) {
	infile := "testdata/set1.fa"
	s_opts := &seq.Options{Vbsty: 3, Keep_gaps: true}
	seq_map := make(map[string]int)
	seq_set := make([]seq.Seq, 0, 0)

	seqs, _, err := seq.ReadSeqs(infile, seq_set[:], seq_map, s_opts)
	if err != nil {
		t.Errorf(err.Error())
	}
	const refSeqNdx = 1
	if err = Calc(seqs, refSeqNdx); err != nil {
		t.Errorf("Bad symbol not recognised")
	}

}	

func TestSeqTypes (t *testing.T) {
	type barray []byte
//  x := [2]seq.Seq{}
//	seqs := x[:]
	var seqs []seq.Seq
	dnaseqs := []barray {[]byte("acgt-acgt"), []byte("ac -ggga ")}
	for _, d := range dnaseqs {
		var t seq.Seq
		t.SetSeq(d)
		seqs = append (seqs, t)
	}
	fmt.Println ("lookatseqs", seqs)
}
