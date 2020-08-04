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

func TestCount (t *testing.T) {
	var fname string
	var err error
	if fname, err = makeTestData(); err != nil {
		t.Fatal(err)
	}
	defer os.Remove (fname)
	type nfunc func (string) (int, error)
	for _, ff := range []nfunc {numseq.ByReadingFixed, numseq.ByMmap, numseq.Main} {
		if i, err := ff (fname); i != smalltestArg.Nseq {
			t.Fatal ("Expected", smalltestArg.Nseq, "got", i)
		} else if err != nil {
			t.Fatal (err)
		}
	}
}

type fToBench func (string) (int, error)
func dobench (b *testing.B, f fToBench) {
	b.StopTimer()
	var fname string
	var err error
	if fname, err = makeTestData(); err != nil {
		b.Fatal(err)
	}
	b.Cleanup (func () {os.Remove(fname)})
	b.StartTimer()
    i, _ := f (fname)
	if i != smalltestArg.Nseq	 {
		b.Fatal ("Expected", smalltestArg.Nseq, "got", i)
	}
}

func BenchmarkByMmap (b *testing.B) {
    dobench (b, numseq.ByMmap)
}

func BenchmarkByReadingFixed (b *testing.B) {
    dobench (b, numseq.ByReadingFixed)
}
