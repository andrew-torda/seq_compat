// This is a second version of splitting lines at spaces and quotes.
// This should split a byte slice at spaces and quotes.

/* from https://www.iucr.org/resources/cif/spec/version1.1/cifsyntax
               character or string role
_ (underscore) identifies data name
#              identifies comment
$              identifies save frame pointer
'              delimits non-simple data values
"              delimits non-simple data values
[              reserved opening delimiter for non-simple data values (see paragraph 19)
]              reserved closing delimiter for non-simple data values (see paragraph 19)
; at beginning of line of text delimits non-simple data values
data_          identifies data block header (case-insensitive)
save_          identifies save frame header or terminator (case-insensitive)
*/

package mmcif

import (
	"errors"
)

// fields takes a string and breaks it into a slice of strings which
// are the space separated words. Unlike the library version, it
// take a slice as an argument and fills it out. If the slice is not big
// enough, fields will be lost.
// It is in the middle of a loop and it was
// one of the main causes of memory allocations and an expensive part
// of the code..
//BenchmarkLibFields-4   	 1000000	      2585 ns/op using library call
//BenchmarkFields1-4     	 2000000	       656 ns/op stripped down version
//BenchmarkFields2-4     	 3000000	       443 ns/op version below

func fields(s bSlice, scrtch []bSlice) []bSlice {
	var i, istart, iwrd int

	for i = 0; i < len(s); i++ { // leading spaces
		if iswhite(s[i]) {
			continue
		}
		break
	}

	if i == len(s) || (cap(scrtch) == 0) {
		return nil
	}
	istart = i
	for {
		for { //                   in a word
			if iswhite(s[i]) {
				scrtch[iwrd] = s[istart:i]
				iwrd++
				if int(iwrd) == cap(scrtch) {
					return scrtch[0:iwrd]
				}
				break
			}
			i++
			if i == len(s) {
				scrtch[iwrd] = s[istart:i]
				return scrtch[0 : iwrd+1]
			}
		}
		for i++; ; i++ { //        in spaces
			if i == len(s) {
				return scrtch[0:iwrd]
			}
			if !iswhite(s[i]) {
				break
			}
		}
		istart = i
	}
}

// iswhite only works for ascii spaces
var asciiSpace = [256]bool{
	'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true,
}

// iswhite returns true if a byte is on the list of white space characters.
func iswhite(b byte) bool {
	return asciiSpace[b] // Seems to be inlined, so it costs nothing.
}

// isquote not only checks if we have a quote character, but also
func isquote(b byte, qtype *byte) bool { // stores its type
	if b == squote || b == dquote { //     (single or double) so we can
		*qtype = b  //                      so we can look
		return true //                      for the corresponding
	} //                                    closing quote
	return false
}

type sInfo struct { // Holds the state of the state functions
	err     error
	ret     []([]byte) //This is what we will really return
	byteIn  []byte
	nxtIndx int
	qtype   byte // type of quote
}
type sfn func(i int, c byte, s *sInfo) sfn // state function

func sfnInQuote(i int, c byte, sInfo *sInfo) sfn { // First state, in quoted region
	if c == sInfo.qtype {
		return sfnExitQuote
	}
	if c == '\n' {
		sInfo.err = errors.New("unterminated quote line: " + string(sInfo.byteIn))
		return sfnWhite
	}
	return sfnInQuote
}

func sfnExitQuote(i int, c byte, sInfo *sInfo) sfn { // Second state
	if iswhite(c) { // quote followed by white really ends a quoted region
		t := sInfo.byteIn[sInfo.nxtIndx : i-1]
		sInfo.ret = append(sInfo.ret, t)
		return sfnWhite
	}
	return sfnInQuote // but if a character comes, we go back to quoted region
}

func sfnInText(i int, c byte, sInfo *sInfo) sfn {
	if iswhite(c) {
		t := sInfo.byteIn[sInfo.nxtIndx:i]
		sInfo.ret = append(sInfo.ret, t)
		return sfnWhite
	}
	return sfnInText
}

func sfnWhite(i int, c byte, sInfo *sInfo) sfn { // State - in white space region
	switch {
	case iswhite(c):
		return sfnWhite
	case isquote(c, &sInfo.qtype):
		sInfo.nxtIndx = i + 1
		return sfnInQuote
	default:
		sInfo.nxtIndx = i
		return sfnInText
	}
}

// splitCifLine takes a byte slice and returns a set byte slices consisting
// of words from the original slice. They are separated by spaces and matching
// quotes.
// We have a small finite state machine with four states. When we leave text or
// a quote followed by a space, we save the word and append it to "ret"
//
//
func splitCifLine(byteIn []byte, retIn [][]byte) ([]([]byte), error) {
	if len(byteIn) < 1 {
		return nil, nil
	}

	var sInfo = sInfo{ret: retIn[:0], byteIn: byteIn} // costs memory

	state := sfnWhite
	for i, c := range byteIn {
		state = state(i, c, &sInfo) // escapes here
	}
	state(len(byteIn), '\n', &sInfo) // end with newline, catches unterminated quotes
	if sInfo.err != nil {            // Just check at end, to avoid if statements within loop
		return nil, sInfo.err
	}
	return sInfo.ret, nil
}
