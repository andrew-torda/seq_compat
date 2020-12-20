// 27 April 2020

package entropy_test

import (
	. "github.com/andrew-torda/seq_compat/pkg/entropy"
	"github.com/andrew-torda/seq_compat/pkg/seq"
	"github.com/andrew-torda/seq_compat/pkg/seq/common"
	"math"
	"os"
	"testing"
)

// The real testing for the calculations is in the directory with the
// entropy calculation.

// Write some junk to a file, make sure main opens it and does the right
// thing. Return the name of the file.

var seqstring string = `>s1
ACGT
> s2
-CGT
> s3
-GGT`

// check case where there are no gaps
var seqstring2 string = `> s1
ACGT
>s2
ACGT`

var seqstring3 string = `> s1
aaaaa
> s2
aaa-c
> s3
aa--d
> s4
a---e`

func ExampleMain(t *testing.T) {
	var fname string
	var err error
	runs := []string{seqstring, seqstring2}
	for _, s := range runs {
		if fname, err = common.WrtTemp(s); err != nil {
			t.Fatal("Fail writing test file")
		}
		defer os.Remove(fname)
		flags := CmdFlag{}
		if err := Mymain(&flags, fname, ""); err != nil {
			t.Fatal("bust on simple test", err)
		}
	}
	// Output:
	// boo
}

// approxEqual
func approxEqual(x, y float32) bool {
	const eps = 0.000001
	d := x - y
	if d > eps || d < -eps {
		return false
	}
	return true
}

// Test2
func Test2(t *testing.T) {
	log4 := func(x float64) float64 { return math.Log(x) / math.Log(4.) }
	ss := [][]string{
		{"aaaa", "abab", "acbb", "adbb"},
	}

	x3of4 := -float32((3./4)*(log4(3./4)) - 1./4)
	wantEnt := []float32{0, 1, 0.5, x3of4}
	seqgrp := seq.Str2SeqGrp(ss[0], "tt0")
	if seqgrp.GetNSym() != 4 {
		t.Fatal("Test2 not written correctly")
	}
	gapsAreChar := false
	gotEnt := make([]float32, seqgrp.GetLen())
	seqgrp.Entropy(gapsAreChar, gotEnt)

	for i := range gotEnt {
		if gotEnt[i] != wantEnt[i] {
			t.Fatal("entropy i", i, "got", gotEnt[i], "want", wantEnt[i])
		}
	}
}

// testChimera see what happens if we write a chimera format file
/*  This is commented out while I try to think of a way to test the output.
func TestChimera(t *testing.T) {
	var fname string
	var err error
	if fname, err = common.WrtTemp(seqstring3); err != nil {
		t.Fatal("Fail writing test file")
	}
	defer os.Remove(fname)
	var tmpoutName string
	if tmpout, err := ioutil.TempFile("", "del_me"); err != nil {
		t.Fatal("Fail making test file", err.Error())
	} else {
		tmpoutName = tmpout.Name()
		defer os.Remove(tmpoutName)
		tmpout.Close()
		fmt.Println ("tmpout is ", tmpoutName)
	}
	flags := CmdFlag{
		Chimera: tmpoutName,
	}
	if err := Mymain(&flags, fname, tmpoutName); err != nil {
		t.Fatal("bust with chimera file", err)
	}
}
*/
