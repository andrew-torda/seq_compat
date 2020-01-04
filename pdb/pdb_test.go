package pdb_test

import (
	"os"
	"path/filepath"
	. "pdb"
	"pdb/cmmn"
	"pdb/mmcif"
	"testing"
)

var testdir = "mmcif/testdata/"

// TestBrokenFile checks if we get sensible error messages when we open
// something that is not an mmcif file.
func TestBrokenFile(t *testing.T) {
	ftype := cmmn.FileSrc
	testfiles := []string{
		"/proc",
		"/does/not/exist",
		"/dev/zero",
		os.Args[0],
	}
	for _, s := range testfiles {
		chains, err := ReadCoord(s, ftype, "stdout")
		if chains != nil {
			t.Error("chains should be nil")
		}

		if err == nil {
			t.Error("Did not get expected error on", s)
		}
	}
}

//
func TestNatomstot(t *testing.T) {
	testdir = "mmcif/testdata/"
	testfile := filepath.FromSlash(testdir + "threemodel.cif")
	fp, err := os.Open(testfile)
	if err != nil {
		t.Error("broke opening", testfile)
	}
	defer fp.Close()
	mr := mmcif.NewMmcifReader(fp)
	mr.SetChains([]string{""})
	mr.SetModelMax(3)
	mr.SetAtoms([]string{"CA", "CB", "N"})
	md, e2 := mr.DoFile()
	if e2 != nil {
		t.Error("file:", testfile, e2)
	}
	mr = nil
	ch := md.GetChains()
	md = nil
	jValid, jInvalid := NatomsTot(ch)
	print(jValid)
	print(jInvalid)
}

var fnameTypes = []struct {
	fname string
	ftype byte
}{
	{"boo.mmcif", Mmcif_fmt},
	{"boo.mmcif.gz", Mmcif_fmt},
	{"a/b/c.ent", Old_fmt},
	{"a\\b.ent.gz", Old_fmt},
	{"a.pdb", Old_fmt},
	{"a.pdb.gz", Old_fmt},
	{"testdata/ememcif1.gz", Mmcif_fmt},
	{"testdata/ememcif2", Mmcif_fmt},
	{"testdata/peedeebee1", Old_fmt},
	{"testdata/peedeebee2", Old_fmt},
}

func TestOldOrMmcif(t *testing.T) {
	for _, f := range fnameTypes {
		r, err := OldOrMmcif(f.fname)
		if err != nil {
			t.Error("unexpected problem in ", t.Name())
		}
		if r != f.ftype {
			t.Error("in", t.Name(), "working on ", f.fname)
		}
	}
}
