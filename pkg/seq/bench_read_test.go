// Check the effect of different read buffer sizes

package seq_test

import (
	"strings"
	"testing"

	"github.com/andrew-torda/seq_compat/pkg/randseq"
	"github.com/andrew-torda/seq_compat/pkg/seq"
)

func benchmarkReadFasta(i int, b *testing.B) {
	b.ReportAllocs()
	b.StopTimer()

	var sb strings.Builder
	args := randseq.RandSeqArgs{
		Wrtr: &sb,
		Cmmt: "testing seq",
		Nseq: 2000,
		Len:  1937,
	}
	if err := randseq.RandSeqMain(&args); err != nil {
		b.Fatal(err)
	}

	seq.SetFastaRdSize(i)

	reader := strings.NewReader(sb.String())
	s_opts := &seq.Options{
		Keep_gaps_rd: false,
		Dry_run:      true,
		Rmv_gaps_wrt: true,
	}

	var seqgrp, junk seq.SeqGrp
	_ = seq.ReadFasta(strings.NewReader(sb.String()), &junk, s_opts)

	

	b.StartTimer()
	if err := seq.ReadFasta(reader, &seqgrp, s_opts); err != nil {
		b.Fatal("Reading seqs failed", err)
	}

	if seqgrp.GetNSeq() != args.Nseq {
		b.Fatalf("Got %d wanted %d seqlen was %d\n", seqgrp.GetNSeq(), args.Nseq, seqgrp.GetLen())
	}
}

func Benchmark3(b *testing.B)   { benchmarkReadFasta(3, b) }
func Benchmark4(b *testing.B)   { benchmarkReadFasta(4, b) }
func Benchmark512(b *testing.B) { benchmarkReadFasta(512, b) }
func Benchmark2k(b *testing.B)  { benchmarkReadFasta(2*1024, b) }
func Benchmark4k(b *testing.B)  { benchmarkReadFasta(4*1024, b) }
func Benchmark10k(b *testing.B) { benchmarkReadFasta(10*1024, b) }
func Benchmark20k(b *testing.B) { benchmarkReadFasta(20*1024, b) }
func Benchmark40k(b *testing.B) { benchmarkReadFasta(40*1024, b) }
