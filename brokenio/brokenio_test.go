package brokenio_test

import (
	"github.com/andrew-torda/goutil/brokenio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var tochop = [][]byte{
	[]byte(""),
	[]byte("a"),
	[]byte("abc"),
	[]byte("abcdefghij"),
	[]byte("abcdefghijklmn"),
}

var longstring = "0123456789012345678901234567890123456789"

// len_non_null returns the length of byte array up to first null
func lenNonNull(a []byte) int {
	for i := 0; i < len(a); i++ {
		if a[i] == 0 {
			return i
		}
	}
	return len(a)
}

// checkNonNull gets two byte slices and sees if they are
// identical within the first characters which are not nulls
func checkNonNull(a, b []byte) bool {
	shorter := lenNonNull(a)
	if x := lenNonNull(b); x < shorter {
		shorter = x
	}
	if bytes.Equal(a[:shorter], b[:shorter]) {
		return true
	}
	return false
}

// testFrac - wipe out different fractions of the input buffer.
func testFrac(t *testing.T, inb []byte, frac float32) {
	s := make([]byte, len(inb))
	rdr := brokenio.NewReader(ioutil.NopCloser(strings.NewReader(string(inb))))
	rdr.SetProbFail(1)
	rdr.SetFracFail(frac)
	_, err := rdr.Read(s) // Look at the error in the different cases below
	nuls := []byte{0}
	if checkNonNull(inb, s) == false {
		t.Error("contents of strings changed with string", string(inb), "frac", frac)
	}
	switch frac {
	case 0.0:
		if n := bytes.Count(s, nuls); n > 0 {
			t.Error("want no null bytes, got", n)
		}
		if err != nil && len(inb) > 0 {
			t.Errorf("error reading from string \"%s\"", inb)
		}
	case 1.0: // This should be a string with all nulls and an error
		want := len(s)
		if n := bytes.Count(s, nuls); n != want {
			t.Error("want", want, "nulls, got", n)
		}
		if len(s) > 0 && err == nil {
			t.Error("did not get error reading from", string(inb))
		}
	default:
		nNull := bytes.Count(s, nuls)
		if nNull == 0 && len(inb) > 0 {
			t.Errorf("no nulls found in \"%s\"", string(s))
		}
		if nNull == len(s) && len(s) > 2 {
			t.Error("Wiped out complete string in", string(inb))
		}
	}
	//	fmt.Println("frac=", frac, "inb:", string(inb), "s", string(s))
}

// TestTrashing takes strings and removes parts of them
func TestTrashing(t *testing.T) {
	fracs := [3]float32{0, 0.3, 1}
	for _, frac := range fracs {
		for _, inb := range tochop {
			testFrac(t, inb, frac)
		}
	}
}

func forZeroFile(prob float32) (n int, err error) {
	rdr := brokenio.NewReader(ioutil.NopCloser(strings.NewReader(longstring)))
	rdr.SetProbZeroFile(prob)
	tmp := make([]byte, len(longstring))
	n, err = rdr.Read(tmp)
	rdr.Close()
	return n, err
}

func TestZeroFile(t *testing.T) {
	n, err := forZeroFile(1)
	if n > 0 {
		t.Error("should have received zero bytes")
	}
	if err != io.EOF {
		t.Errorf("Should have recieved EOF")
	}
	n, err = forZeroFile(0)
	if n < len(longstring) {
		t.Error("Wanted", len(longstring), "got", n)
	}
	if err != nil {
		t.Errorf("err reading from string")
	}
}

func TestReaderSimple(t *testing.T) {
	rdr := brokenio.NewReader(ioutil.NopCloser(strings.NewReader(longstring)))
	rdr.SetProbFail(0)
	s := make([]byte, len(longstring))
	if rdr.Read(s); string(s) != longstring {
		t.Errorf("simple read fail got %q wanted %q", s, longstring)
	}
}

func Example_setVerbose() {
	rdr := brokenio.NewReader(ioutil.NopCloser(strings.NewReader(longstring)))
	rdr.SetVerbose(true)
	tmp := make([]byte, len(longstring))
	rdr.Read(tmp)
	rdr.Close()
	// Output: Closing 1 calls and 40 bytes
}

// TestClose - check if the reader really is calling the correct close method.
// It is.
func TestClose(t *testing.T) {
	dir := os.TempDir()
	f, err := ioutil.TempFile(dir, "testclose_test")
	if f == nil || err != nil {
		t.Errorf("TempFile(dir, testclose_test) = %v, %v", f, err)
	}
	defer os.Remove(f.Name())

	if n, err := f.Write([]byte(longstring)); n != len(longstring) {
		t.Error("Writing temp file failed", err)
	}
	f.Close()
	fp, err := os.Open(f.Name())
	if fp == nil || err != nil {
		t.Error("reading from tempfile, err = ", err)
	}
	rdr := brokenio.NewReader(fp)
	s := make([]byte, len(longstring))
	if n, err := rdr.Read(s); n != len(longstring) || err != nil {
		t.Error("Failed reading from tempfile, n, err = ", n, err)
	}
	rdr.SetVerbose(false)
	if err = rdr.Close(); err != nil {
		t.Error("failed on close of reader")
	}
}
