// 30 May 2020

// Do a test where we re-use the same sequences. Check for complete
// conservation and completely different.
package kl_test

import (
	"io/ioutil"
	"math"
	"os"
	"testing"

	. "github.com/andrew-torda/seq_compat/pkg/kl"
	"github.com/andrew-torda/seq_compat/pkg/randseq"
	"github.com/andrew-torda/seq_compat/pkg/seq"
)

// TestTiny - just make sure we can open and read some files.
func TestTiny(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "delete_me")
	if err != nil {
		t.Fatalf("Broke on tempfile %v", err)
	}
	cleanup := func() {
		os.Remove(tmpfile.Name())
	}
	t.Cleanup(cleanup)
	var args randseq.RandSeqArgs
	args = randseq.RandSeqArgs{Wrtr: tmpfile, Iseed: 1, Len: 5, Nseq: 6}
	randseq.RandSeqMain(&args)
	tmpfile.Close()
	var flags CmdFlag
	if err = Mymain(&flags, tmpfile.Name(), tmpfile.Name(), "boo"); err != nil {
		t.Fatalf("broke on tiny set: %v", err)
	}
}

// TestError1 exercises code for broken files. It is not silly, since
// there are gymnastics to handle errors from two threads.
func TestError1(t *testing.T) {
	var flags CmdFlag
	shouldProvoke := "should have provoked an error message"
	for i := 0; i < 30; i++ { // there are some channels involved, so repeat
		err := Mymain(&flags, "notexist", "testdata/a.fa", os.DevNull)
		if err == nil { // first file is missing
			t.Error(shouldProvoke)
		}
		err = Mymain(&flags, "testdata/a.fa", "/notexist", os.DevNull)
		if err == nil { // second file is missing
			t.Error(shouldProvoke)
			err = Mymain(&flags, "/notexist_1", "/notexist_2", os.DevNull)
			if err == nil { // both files are missing
				t.Error(shouldProvoke)
			}
		}
	}
}

func TestSimple1(t *testing.T) {
	var flags CmdFlag
	s1 := []string{"aa", "aa"}
	s2 := []string{"ab", "bb"}
	seqgrp1 := seq.Str2SeqGrp(s1, "s")
	seqgrp2 := seq.Str2SeqGrp(s2, "t")
	var seqX1, seqX2 SeqX

	if err := ExtractSeqX(seqgrp1, &seqX1, &flags); err != nil {
		t.Fatal(err)
	}
	if err := ExtractSeqX(seqgrp2, &seqX2, &flags); err != nil {
		t.Fatal(err)
	}
}

func approxEqual(x, y float32) bool {
	d := x - y
	const eps = 0.0001
	if d < -eps || d > eps {
		return false
	}
	return true
}

// f to save typing
func f(want, got float32, i int, name string, t *testing.T) {
	const bust = "%s wanted %f got %f column %d"
	if !approxEqual(want, got) {
		t.Fatalf(bust, name, want, got, i)
	}
}
func breaker() {}

// TestKL is for the Kullbach-Leibler distance. It is a bit difficult since
// the test columns run across the page.
func TestKL(t *testing.T) {
	// p log p/q
	var messyA, messyB, messyC float32
	{
		log4 := math.Log(4)
		p1a := 1.      // a frac in column
		p2a := 1. / 2. // b frac
		p1b := 1. / 5. // a frac in column
		p2b := 1. / 2. // b frac in column using pseudocounts
		a := p1a * (math.Log(p1a/p2a) / log4)
		b := 0 * (math.Log(p1b/p2b) / log4) // do not use pseudocounts here

		messyA = float32(a + b)
		a = p2a * (math.Log(p2a/p1a) / log4)
		b = p2b * (math.Log(p2b/p1b) / log4)
		messyB = float32(a + b)

		messyC = float32(math.Log(5.) / log4)
	}
	ss := [][]string{
		{"aaaaa", "bbaaa", "ccaaa", "dcaaa"},
		{"aaaab", "bbaab", "ccabb", "dcabb"},
	}

	kl_r := []float32{0, 0, 0, messyA, messyC}
	kl_s := []float32{0, 0, 0, messyB, messyC}
	ent_r := []float32{1, 3. / 4, 0, 0, 0}
	ent_s := []float32{1, 3. / 4, 0, 1. / 2, 0}
	cosS := []float32{1, 1, 1, float32(1. / math.Sqrt(2.)), 0}
	var flags CmdFlag
	seqgrp1 := seq.Str2SeqGrp(ss[0], "tt0")
	seqgrp2 := seq.Str2SeqGrp(ss[1], "tt1")

	var seqX1, seqX2 SeqX
	if err := ExtractSeqX(seqgrp1, &seqX1, &flags); err != nil {
		t.Fatal(err)
	}
	if err := ExtractSeqX(seqgrp2, &seqX2, &flags); err != nil {
		t.Fatal(err)
	}

	klP, klQ, entropyP, entropyQ, cosSim := CalcInner(seqX1, seqX2)
	for i := range klP {
		f(kl_r[i], klP[i], i, "klP", t)
		f(kl_s[i], klQ[i], i, "klQ", t)

		f(ent_r[i], entropyP[i], i, "entropy P", t) // race
		f(ent_s[i], entropyQ[i], i, "entropy Q", t)
		f(cosS[i], cosSim[i], i, "cosine sim", t)
	}
}

// TestInnerCosSim takes a few simple vectors with their cosine similarity,
// but it does a few extra checks. The cosine similarity should be independent
// of vector length, so we scale the vectors and check that the answer does
// not change. The answer should also be the same if we permute both vectors.
func checkCosSim(v1, v2 []float32, r float32, t *testing.T) {
	got := InnerCosSim(v1, v2)
	if got != r {
		t.Fatal("CosSim got", got, "wanted", r, "vec 1", v1, "vec 2", v2)
	}
}
func scalevec(v []float32, scale float32) []float32 {
	for i := range v {
		v[i] *= scale
	}
	return v
}

// permuteVec rotates left - moves first element to end
func permuteVec(v []float32) []float32 { return (append(v[1:len(v)], v[0])) }
func TestInnerCosSim(t *testing.T) {
	type f32 []float32

	vecs := []struct {
		v1 []float32
		v2 []float32
		r  float32
	}{
		{f32{0, 1}, f32{0, 1}, 1},
		{f32{0, 1, 0}, f32{0, 1, 0}, 1},
		{f32{0, 1}, f32{1, 0}, 0},
		{f32{1, 0}, f32{1, 1}, float32(math.Sqrt(2) / 2)},
		{f32{5, 4, 3, 2, 1}, f32{5, 4, 3, 2, 1}, 1},
	}

	for _, s := range vecs {
		checkCosSim(s.v1, s.v2, s.r, t)
		checkCosSim(s.v1, scalevec(s.v2, 17.3), s.r, t)
		checkCosSim(permuteVec(s.v1), permuteVec(s.v2), s.r, t)
		checkCosSim(scalevec(s.v1, -1), scalevec(s.v2, -1), s.r, t)
		checkCosSim(scalevec(s.v1, -1), scalevec(s.v2, 1), -s.r, t)

	}
}
