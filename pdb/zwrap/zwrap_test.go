// Test Zwrap
package zwrap_test

import (
	"github.com/andrew-torda/goutil/pdb/zwrap"
	"errors"
	"io"
	"io/ioutil"
	"testing"
	"os"
)

// both of these are "andrewsays", but the first is compressed. Write them to a file
// and check that the file opener does the right thing.
type gztest struct {
	data    []byte
	gzipped bool
}

var gztests = []gztest{
	{[]byte{
		0x1f, 0x8b, 0x08, 0x00, 0xb6, 0xf1, 0xa0, 0x5b, 0x00, 0x03,
		0x4b, 0xcc, 0x4b, 0x29, 0x4a, 0x2d, 0x2f, 0x4e, 0xac, 0x2c,
		0xce, 0x48, 0xcd, 0xc9, 0xc9, 0x07, 0x00, 0x44, 0xa8, 0x66,
		0x89, 0x0f, 0x00, 0x00, 0x00},
		true,
	},
	{[]byte{
		0x61, 0x6e, 0x64, 0x72, 0x65, 0x77, 0x73, 0x61,
		0x79, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x0a},
		false,
	},
}

// writeToTmp writes a bitslice to a temporary file and returns
// a file pointer.
func writeToTmp (data []byte ) (*os.File, error) {
	tmpf, err := ioutil.TempFile("", "del_me_testing")
	if err != nil {
		return nil, errors.New("Fail getting TempFile")
	}
	if _, err := tmpf.Write(data); err != nil {
		return nil, errors.New("fail writing to tempfile")
	}
	if _, err := tmpf.Seek (0, io.SeekStart) ; err != nil {
		return nil, errors.New("Seek fail on " + tmpf.Name())
	}
	return tmpf, nil
}


func TestWrap (t *testing.T) {
	b := make ([]byte, 256)
	for _, x := range gztests {
		tmpfp, err := writeToTmp (x.data)
		if err !=nil {
			t.Error (err)
		}
		tmpr, err := zwrap.Wrap (tmpfp)
		if err != nil {
			if x.gzipped {
				t.Error("Fail on correctly gzipped file")
			}
			continue // It is not gzipped, so move on to next
		} else { // No error
			if ! x.gzipped { // But we should get one
				t.Error ("Fail on not compressed file")
			}
		}
		if n, err := tmpr.Read (b); n < 5 {
			t.Errorf ("Short read of %d bytes, %s", n, err)
		}
		if string(b[:10]) != "andrewsays"[:10] {
			t.Errorf("wrong string: %s", b[:10])
		}
		if err := tmpr.Close(); err != nil {
			t.Errorf("Error closing: %s", err)
		}
		if err := os.Remove(tmpfp.Name()); err != nil {
			t.Errorf("Fail removing temp file")
		}
	}
}

// Calling WrapMaybe should not fail since it guesses if the file
// is compressed or not.
func TestWrapMaybe (t *testing.T) {
	b := make ([]byte, 256)
	for _, x := range gztests {
		tmpfp, err := writeToTmp (x.data)
		if err !=nil {
			t.Error (err)
		}
		tmpr, err := zwrap.WrapMaybe (tmpfp)
		
		if err != nil {
			t.Errorf ("Fail on file where compressed was %v", x.gzipped)
		}
		if n, err := tmpr.Read (b); n < 5 {
			t.Errorf ("Short read of %d bytes, %s", n, err)
		}
		if string(b[:10]) != "andrewsays"[:10] {
			t.Errorf("wrong string: %s", b[:10])
		}
		if err := tmpr.Close(); err != nil {
			t.Errorf("Error closing: %s", err)
		}
		if err := os.Remove(tmpfp.Name()); err != nil {
			t.Errorf("Fail removing temp file")
		}
	}
}
