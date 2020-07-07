// 30 May 2020

// Do a test where we re-use the same sequences. Check for complete
// conservation and completely different.
package kl_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/andrew-torda/goutil/seq"
	. "github.com/andrew-torda/seq_compat/pkg/kl"
)

// TestError1 exercise code for broken files.
func TestError1(t *testing.T) {
	var flags CmdFlag
	shouldProvoke := "should have provoked an error message"
	for i := 0; i < 30; i++ { // there are some channels involved, so repeat
		err := Mymain(&flags, "notexist", "testdata/a.fa", os.DevNull)
		if err == nil {
			t.Error(shouldProvoke)
		}
		err = Mymain(&flags, "testdata/a.fa", "notexist", os.DevNull)
		if err == nil {
			t.Error(shouldProvoke)
		}
	}
}
func breaker() {}
func TestSimple1(t *testing.T) {
	var flags CmdFlag
	s1 := []string{"aa", "aa"}
	s2 := []string{"ab", "bb"}
	seqgrp1 := seq.Str2SeqGrp(s1, "s")
	seqgrp2 := seq.Str2SeqGrp(s2, "t")
	var seqX1, seqX2 SeqX

	if err := GetSeqX(&seqgrp1, &seqX1, &flags); err != nil {
		t.Fatal(err)
	}
	if err := GetSeqX(&seqgrp2, &seqX2, &flags); err != nil {
		t.Fatal(err)
	}
}

func TestKL(t *testing.T) {
	ss := [][]string{
		{"aaaa", "bbaa", "ccab", "dcab"},
		{"aaaa", "bbaa", "ccaa", "dcaa"},
	}
    kl_r := []float32{0, 0, 0, 0}
	var flags CmdFlag
	fmt.Println ("ss0", ss[0])
	fmt.Println ("ss1", ss[1])

	seqgrp1 := seq.Str2SeqGrp(ss[0], "tt0")
	seqgrp2 := seq.Str2SeqGrp(ss[1], "tt1")
	var seqX1, seqX2 SeqX
	if err := GetSeqX(&seqgrp1, &seqX1, &flags); err != nil {
		t.Fatal(err)
	}
	if err := GetSeqX(&seqgrp2, &seqX2, &flags); err != nil {
		t.Fatal(err)
	}

	klP    := make ([]float32, seqX1.GetLen())
	KlFromSeqX (&seqX1, &seqX2, klP)
	for i := range klP {
		if kl_r[i] != klP[i] {
			t.Fatalf ("kl calc wanted %f got %f", kl_r[i], klP[i])
		}
	}
	fmt.Println (klP)
}
