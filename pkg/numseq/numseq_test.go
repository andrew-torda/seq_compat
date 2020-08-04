package numseq_test

import (
	"os"
	"io/ioutil"
	"testing"

	"github.com/andrew-torda/seq_compat/pkg/randseq"
	"github.com/andrew-torda/seq_compat/pkg/numseq"
)

var smalltestArg = randseq.RandSeqArgs{
	Cmmt: "test seq",
	Nseq: 10000,
	Len:  2000,
}

func makeTestData() (string, error) {
	args := smalltestArg
	f_tmp, err := ioutil.TempFile("", "_del_me_testing")
	if err != nil {
		return "", err
	} else {
		args.Wrtr = f_tmp
		defer f_tmp.Close()
	}

	if err := randseq.RandSeqMain(&args); err != nil {
		return "", err
	}
	return f_tmp.Name(), nil
}

func TestByreading (t *testing.T) {
	var fname string
	var err error
	if fname, err = makeTestData(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove (fname)
	if i, err := numseq.ByreadingFixed (fname, 1); i != smalltestArg.Nseq {
		t.Fatal ("Expected", smalltestArg.Nseq, "got", i)
	} else if err != nil {
		t.Fatal (err)
	}
}

func setupbmark (b *testing.B) (string, int) {
	b.StopTimer()
	var fname string
	var err error
	if fname, err = makeTestData(); err != nil {
		b.Fatal(err)
	}
	b.Cleanup (func () {os.Remove(fname)})
	b.StartTimer()
	return fname, smalltestArg.Nseq
}	

func BenchmarkByreadingFixed(b *testing.B) {
	fname, nset := setupbmark(b)
    i, _ := numseq.ByreadingFixed (fname, 1)
	if i != nset {
		b.Fatal ("Expected", smalltestArg.Nseq, "got", i)
	}
}

func BenchmarkByreadingVary20k(b *testing.B) {
	fname, nset := setupbmark(b)
    i, _ := numseq.ByreadingVaries (fname, 20 * 1024)
	if i != nset {
		b.Fatal ("Expected", nset, "got", i)
	}
}

func BenchmarkByreadingVary50k(b *testing.B) {
	fname, nset := setupbmark(b)
    i, _ := numseq.ByreadingVaries (fname, 50 * 1024)
	if i != nset {
		b.Fatal ("Expected", nset, "got", i)
	}
}

func BenchmarkByreadingVary100k(b *testing.B) {
	fname, nset := setupbmark(b)
    i, _ := numseq.ByreadingVaries (fname, 100 * 1024)
	if i != nset {
		b.Fatal ("Expected", nset, "got", i)
	}
}


func BenchmarkByreadingVary1m(b *testing.B) {
	fname, nset := setupbmark(b)
    i, _ := numseq.ByreadingVaries (fname, 1024 * 1024)
	if i != nset {
		b.Fatal ("Expected", nset, "got", i)
	}
}

func BenchmarkByMmap (b *testing.B) {
	fname, nset := setupbmark(b)
    i, _ := numseq.ByMmap (fname)
	if i != nset {
		b.Fatal ("Expected", nset, "got", i)
	}
}

