// Check the effect of different read buffer sizes

package seq_test

import (
	"strings"
	"testing"

	
	"github.com/andrew-torda/seq_compat/pkg/randseq"
	"github.com/andrew-torda/seq_compat/pkg/seq"
	
)

func benchmarkReadFasta (i int, b *testing.B, oldvers bool) {
	b.ReportAllocs()
	b.StopTimer()

	var sb strings.Builder
	args := randseq.RandSeqArgs {
		Wrtr: &sb,
		Cmmt: "testing seq",
		Nseq: 1000,
		Len: 1600,
	}
	if err := randseq.RandSeqMain(&args); err != nil {
		b.Fatal(err)
	}

	seq.SetFastaRdSize (i)


	reader := strings.NewReader(sb.String())
	s_opts := &seq.Options{
		Vbsty: 0, Keep_gaps_rd: false,
		Dry_run:      true,
		Rmv_gaps_wrt: true,
	}

	var seqgrp, junk seq.SeqGrp
	_ = seq.ReadFasta (strings.NewReader(sb.String()), &junk,  s_opts)

	f := seq.ReadSeqs
	if (oldvers == true) {
		f = seq.ReadSeqs
	}
	b.StartTimer()
	if err := f(reader, &seqgrp, s_opts); err != nil {
		b.Fatal("Reading seqs failed", err)
	}

	if seqgrp.GetNSeq() != args.Nseq  { b.Fatalf ("Got %d wanted %d seqlen was %d\n", seqgrp.GetNSeq(), args.Nseq, seqgrp.GetLen())}
}

func Benchmark3 (b *testing.B) { benchmarkReadFasta (3, b, false) }
func Benchmark4 (b *testing.B) { benchmarkReadFasta (4, b, false) }
func Benchmark512 (b *testing.B) { benchmarkReadFasta (512, b, false) }
func Benchmark4k (b *testing.B) { benchmarkReadFasta (4 * 1024, b, false) }
func Benchmark10k (b *testing.B) { benchmarkReadFasta (10 * 1024, b, false) }
func Benchmark20k (b *testing.B) { benchmarkReadFasta (20 * 1024, b, false) }
func BenchmarkOld (b *testing.B) { benchmarkReadFasta (4 * 1024, b, true) }
