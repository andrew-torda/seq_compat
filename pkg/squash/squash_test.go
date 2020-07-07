 // 29 Apr 2020

package squash_test

import (
	"os"
	"io/ioutil"
	"log"
	"testing"

	"github.com/andrew-torda/goutil/seq/common"
	. "github.com/andrew-torda/seq_compat/pkg/squash"
)

var seqstring string = `>s1
ABCD
> s2
-EFG
> s3
-HIJ`

//func TestMain1(t *testing.T) {
func ExampleMain () {
	fname, err := common.WrtTemp(seqstring)
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(fname)
	if MyMain("s2", fname, "") != common.ExitSuccess {
		log.Fatal("broke running squash main")
	}
// Output:
//>s1
//BCD
//> s2
//EFG
//> s3
//HIJ
}

func TestWithOutput (t *testing.T) {
	fname, err := common.WrtTemp(seqstring)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fname)
	tfile, err := ioutil.TempFile ("", "_del_me_testing")
	outfname := tfile.Name()
	tfile.Close()
	if (err != nil) {
		t.Fatal("cannot make temp filename", err)
	}
	r := MyMain("s2", fname, outfname)
	defer os.Remove(outfname)
	if r != common.ExitSuccess {
		t.Fatal("broke running squash main")
	}
	if fi, err := os.Stat (outfname); err != nil {
		t.Fatal("stat failed", err)
	} else {
		const sOf = "size of output from MyMain is too"
		if fi.Size() < 25 {
			t.Fatal(sOf, "small")
		}
		if fi.Size() > 27 {
			t.Fatal(sOf, "big")
		}
	}
}

// TestBreak checks that we get an error if sequences have
// wrong lengths
func TestBreak(t *testing.T) {
	var seqstring string = `>s1
ABCD
> s2
-EF
> s3
-HIJ`

	fname, err := common.WrtTemp(seqstring)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fname)
	old := os.Stderr // We provoke an error, so temporarily redirect stderr.
	os.Stderr, err = os.OpenFile(os.DevNull, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if (err != nil) {t.Fatal("fail opening", os.DevNull)}
	if MyMain("s2", fname, "") != common.ExitFailure {
		os.Stderr = old
		t.Fatal("broke running squash main")
	}
	os.Stderr = old
}
