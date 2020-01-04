// brokenio is a wrapper around an io.ReadCloser. It allows us to set
// rates of failed read operations.
// Typical use: You get a file pointer, a reader from a compressed
// source or an http source. You write
// reader = newReader(reader) to wrap the old reader. Everything then
// functions as before, but with artificial errors.
// When we introduce an error, we return an error.
// When we introduce a failure on the first read, we return without an
// error. This is what one often sees on a zero length file.

package brokenio

import (
	"fmt"
	"io"
	"math/rand"
)

// A Reader is modelled on the various Readers in the standard library,
// but with variables controlling the frequency of errors.
// These values are the fraction of time an error will take place,
// so a value of 0.05 means failure in 5% of the cases.
// If verbose is true, print out the amount of data when the file is closed.
type BrknRdrClsr struct {
	rdr_orig     io.ReadCloser // Wrapped reader
	probZeroFile float32       // Probability of returning a zero length file
	probFail     float32
	fracFail     float32
	nCalled      int
	nByte        int
	verbose      bool
}

// dfltReader sets default values for a new brokenio reader.
var dfltReader = BrknRdrClsr {
	rdr_orig:     nil,
	probZeroFile: 0,
	probFail:     0,
	fracFail:     0.5,
	verbose:      false,
}

// SetVerbose sets the verbosity flag to true or false
func (r *BrknRdrClsr) SetVerbose(newV bool) { r.verbose = newV }

// SetFracFail sets the amount of the bytes which will be trashed
func (r *BrknRdrClsr) SetFracFail (frac float32) { r.fracFail = frac}

// SetProbZeroFile sets the rate at which we simply return 0 bytes on the
// first read. It must be a value from 0 to 1. We do not check if the
// argument is valid.
func (r *BrknRdrClsr) SetProbZeroFile(prob float32) { r.probZeroFile = prob }

// SetProbFail set the probability of a file reading failure.
// It must be between zero and 1.
func (r *BrknRdrClsr) SetProbFail(prob float32) { r.probFail = prob }

// NewReader returns a new Reader - a wrapper around the old one
func NewReader(rIn io.ReadCloser) *BrknRdrClsr {
	var rOut = dfltReader
	rOut.rdr_orig = rIn
	return &rOut
}

// trashslice_devel wipes out characters randomly
// frac is the probability of deleting a character
func trashSlice_devel(p []byte, frac float32) (int, error) {
	delme := make ([]bool, len(p)) // Set up a slice of values
	n_tokeep  := len(p)
	origN := len(p)
	for i := range delme {         // to be deleted
		x := rand.Float32()
		if x < frac {
			delme[i] = true
			n_tokeep--
		}
	}
	if n_tokeep == len(p) { // Nothing to be trashed
		return len(p), nil
	}
	to_ret := make([]byte, n_tokeep)
	for in, out := 0, 0; out < n_tokeep; in++ {
		if delme[in] == false {
			to_ret[out] = p[in]
			out++
		}
	}
	p = p[:len(to_ret)]
	copy (p, to_ret)
	if len(p) != n_tokeep {
		panic(fmt.Sprintf("len(p) %d, n_tokeep %d", len(p), n_tokeep))
	}
	return n_tokeep, fmt.Errorf("randomly deleted %d of %d chars", origN - n_tokeep, origN)
}

// trashSlice wipes out the second part of a slice.
// The amount to wipe out is given by a fraction, so 0.3
// will wipe out the second 30 % of a slice
func trashSlice(p []byte, frac float32) (int, error) {
	nkeep := int(float32(len(p)) * (1. - frac))
	if nkeep == len(p) {
		return nkeep, nil
	}
	err := fmt.Errorf ("randomly wiped out last %d of %d", len(p) - nkeep, len(p))
	q := p[nkeep:]       // Wipe out slice from this point on
	n := len(q)          // How much to wipe out
	x := make([]byte, n) // Guaranteed zero'd data
	copy(q, x)           // Write back to original data.
	return nkeep, err
}

// Read wraps the original reader and sums up the amount of data that
// has gone through. It generates an error with a probability given by freqFailure.
// On the first call, we might return zero data to simulate a zero length file
// which is a rather common occurrence.
func (r *BrknRdrClsr) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.nCalled == 0 {
		if r.probZeroFile > 0 {
			rnd := rand.Float32()
			if (rnd < r.probZeroFile) {
				return 0, io.EOF
			}
		}
	}
	n, err = r.rdr_orig.Read(p)
	r.nCalled++
	r.nByte += n
	rnd := rand.Float32()
	if rnd < r.probFail && r.fracFail > 0 {
		m, err := trashSlice(p, r.fracFail)
		return m, err
	}
	return n, err
}

// Close wraps the original Close method.
func (r *BrknRdrClsr) Close() error {
	if r.verbose {
		fmt.Println("Closing", r.nCalled, "calls and", r.nByte, "bytes")
	}
	return r.rdr_orig.Close()
}
