package numseq_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/andrew-torda/seq_compat/pkg/numseq"
	"github.com/andrew-torda/seq_compat/pkg/randseq"
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

func TestFrompointer(t *testing.T) {
	var fname string
	var err error
	if fname, err = makeTestData(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fname)
	fp, err := os.Open(fname)
	if err != nil {
		t.Fatal("Opening file we just created")
	}
	defer fp.Close()


	if i, err := numseq.ByReading(fp); i != smalltestArg.Nseq {
		t.Fatal("Expected", smalltestArg.Nseq, "got", i)
	} else if err != nil {
		t.Fatal(err)
	}
	fp.Seek(0, io.SeekStart)
	if i, err := numseq.ByMmap(fp); i != smalltestArg.Nseq {
		t.Fatal("Expected", smalltestArg.Nseq, "got", i)
	} else if err != nil {
		t.Fatal(err)
	}



}

type fToBench func(string) (int, error)

func dobench(b *testing.B, f fToBench) {
	b.StopTimer()
	var fname string
	var err error
	if fname, err = makeTestData(); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { os.Remove(fname) })
	b.StartTimer()
	i, _ := f(fname)
	if i != smalltestArg.Nseq {
		b.Fatal("Expected", smalltestArg.Nseq, "got", i)
	}
}

/*
func BenchmarkByMmap (b *testing.B) {
    dobench (b, numseq.ByMmap)
}

func BenchmarkByReadingFixed (b *testing.B) {
    dobench (b, numseq.ByReadingFixed)
}
*/
