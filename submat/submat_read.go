// 23 Feb 2018
// read a substitution matrix
// Next steps
//  * write an inner version that reads from an io.reader. This would let
//    me read from file and strings
//  * move scoring from identity matrix to this file (currently in Gotoh)

package submat

import (
	"github.com/andrew-torda/goutil/matrix"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

// Submat is the export type. it internals do not have to be exported.
type Submat struct {
	mat  *matrix.FMatrix2d
	cmap [128]int8
}

const notset int8 = -1

// String prints out a substitution matrix. Useful during debugging.
func (submat *Submat) String() (s string) {
	cmap := submat.cmap[:]
	s = "Mapping\n"
	n := 10
	for i := range cmap {
		if cmap[i] != notset {
			s = s + fmt.Sprintf("%4s%4d", string(i), cmap[i])
			n--
			if n == 0 {
				n = 10
				s = s + "\n"
			}

		}
	}
	s += "\nThe matrix\n"
	s += fmt.Sprintf("%4s", " ")
	for c := '*'; c < 'Z'; c++ {
		if cmap[c] != notset {
			s += fmt.Sprintf("%4s", string(c))
		}
	}
	s += "\n"
	for c := '*'; c < 'Z'; c++ {
		if cmap[c] != notset {
			s += fmt.Sprintf("%4s", string(c))
			for d := '*'; d < 'Z'; d++ {
				if cmap[d] != notset {
					f := submat.mat.Mat[cmap[c]][cmap[d]]
					s += fmt.Sprintf("%4.0f", f)
				}
			}
			s += "\n"
		}
	}
	return s
}

// CmmtScanner is a wrapper around bufio.Scanner that will ignore anything
// after a comment character and remove leading and trailing white space.
type CmmtScanner struct {
	bufio.Scanner
	cmmt byte // Comment character
}

// NewCmmtScanner is a wrapper around scanner, but
//  - jumps over blank lines
//  - removes leading spaces
//  - removes anything after a comment character
func NewCmmtScanner(r io.Reader, cmmt byte) *CmmtScanner {
	s := bufio.NewScanner(r)
	return &CmmtScanner{*s, cmmt}
}

// CBytes presents exactly the same interface as scanner.Bytes, but
// has to do a bit more work.
// Before returning, we remove anything after the comment symbol and
// strip leading and trailing white space.
// If this leaves us with an empty string, we call Scan again.
// Like the Bytes function, this works directly in the i/o buffer
// and does not allocate any memory. If you like the string it returns,
// you have to save it somewhere.
func (s *CmmtScanner) CBytes() []byte {
	ok := true
	for b := s.Bytes(); ok; ok, b = s.Scan(), s.Bytes() {
		for i := 0; i < len(b); i++ {
			if b[i] == s.cmmt {
				b = b[:i]
				break
			}
		}
		b = bytes.TrimSpace(b)
		if len(b) > 0 {
			return b
		}
	}
	return nil
}

// The first non-comment line  of the substitution matrix file
// contains a list of the allowed characters. Each field has to be
// one character long
func alfbt_line(inline []byte, submat *Submat) (n_alfbt int, err error) {
	cmap := submat.cmap[:]
	for i := range submat.cmap {
		cmap[i] = -1
	}
	f := bytes.Fields(inline)
	n_alfbt = len(f)
	for _, c := range f {
		if len(c) != 1 {
			err = errors.New("alfbt_line: expected a single character, got " + string(c))
			return
		}
		if c[0] > 128 {
			err = errors.New("alfbt_line: saw a non-ascii character in " + string(inline))
			return
		}
	}
	for i, c := range f {
		cmap[c[0]] = int8(i)
	}
	for i, c := range f { // If not set, set both upper and lower case
		l := (bytes.ToLower(c))[0] // This is safe, since we have checked
		u := (bytes.ToUpper(c))[0] // that c is one-byte long
		if cmap[l] == notset {     // Lower case index
			cmap[l] = int8(i)
		}
		if cmap[u] == notset { //     Corresponding upper case index
			cmap[u] = int8(i)
		}
	}

	return len(f), err
}

// Read will read a substitution matrix from a filename.
// Return a pointer to a Submat structure.
func Read(fname string) (submat *Submat, err error) {
	var n_alfbt int
	submat = new(Submat)
	fp, err := os.Open(fname)
	if err != nil {
		return submat, err
	}
	defer fp.Close()
	scnr := NewCmmtScanner(fp, '#')
	scnr.Scan()
	if n_alfbt, err = alfbt_line(scnr.CBytes(), submat); err != nil {
		return submat, err
	}
	submat.mat = matrix.NewFMatrix2d(n_alfbt, n_alfbt)
	r := "Reading from " + fname
	s := ". Wrong number of items on line:\n"
	nc := 0
	for scnr.Scan() {
		line := scnr.CBytes()
		fields := bytes.Fields(line)
		if len(fields) != n_alfbt+1 {
			err = errors.New(r + s + string(line))
			return
		}
		if fields[0][0] > 128 || fields[0][0] < 0 {
			err = errors.New(r + "invalid character on line " + string(line))
			return
		}
		i := submat.cmap[fields[0][0]]
		for j := 0; j < n_alfbt; j++ {
			s := fields[j+1]
			if f, e := strconv.ParseFloat(string(s), 32); err != nil {
				err = errors.New(r + e.Error())
				return
			} else {
				x := float32(f)
				submat.mat.Mat[i][j], submat.mat.Mat[j][i] = x, x
			}
		}
		nc++
	}
	if err = scnr.Err(); err != nil {
		err = errors.New(r + err.Error())
	}
	if nc != n_alfbt {
		err = errors.New(r + ".. not enough lines found")
		return
	}
	return submat, err
}

// Score returns the similarity score of bytes a and b, given
// a specific scoring matrix.
func (submat *Submat) Score(a, b byte) (f float32) {
	i := submat.cmap[a]
	j := submat.cmap[b]
	return submat.mat.Mat[i][j]
}

// ScoreSeqs will take two sequences and calculate a similarity matrix
// based on the substitution matrix.
// We return an M x N matrix, where M and N are the lengths of first
// and second sequences respectively.
func (submat *Submat) ScoreSeqs(s, t []byte) (scr_mat *matrix.FMatrix2d) {
	scr_mat = matrix.NewFMatrix2d(len(s), len(t))
	mat := scr_mat.Mat
	for i, cs := range s {
		for j, ct := range t {
			mat[i][j] = submat.Score(cs, ct)
		}
	}
	return
}
