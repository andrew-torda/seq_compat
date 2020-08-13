// I will put this somewhere else, when I am happy

package white

import (
	"bytes"
	"fmt"
)

// WhiteRemove acts on a byte slice, in place and removes all the white
// space. Note, this changes len().

func Remove(sIn *[]byte) {
	var asciiSpace = [256]bool{
		0: true, '\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true,
	}

	s := *sIn
	i, j := 0, 0
	for ; j < len(s); i, j = i+1, j+1 {
		for j < len(s) {
			if asciiSpace[s[j]] {
				j++
			} else {
				break
			}
		}
		if j >= len(s) {
			break
		}
		s[i] = s[j]
	}

	*sIn = s[:i] // This is the truncation
}

// This version will copy block by block, instead of byte by byte
func removeByBlock(sIn *[]byte) {
	var asciiSpace = [256]bool{
		0: true, '\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true,
	}

	s := *sIn
	inwhite := true
	var start, dst int

	for i := 0; i < len(s); i++ {
		if inwhite {
			if !asciiSpace[s[i]] {
				start = i
				inwhite = false
				continue
			}
		} else {
			if asciiSpace[s[i]] { // end of block
				ltmp := i - start // this is the correct length
				copy(s[dst:dst+ltmp], s[start:i+1])
				dst = dst + ltmp
				inwhite = true
			}
		}
	}

	if !inwhite { // hit end of input, so spit out leftovers
		ltmp := len(s) - start
		s = append(s[:dst], s[start:start+ltmp]...)

	} else { // If we finished in white space, we have to truncate
		s = s[:dst]
	}
	*sIn = s
}

func RemoveWithGoTo(sIn *[]byte) {
	var asciiSpace = [256]bool{
		0: true, '\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true,
	}

	s := *sIn
	fmt.Printf("In: \"%s\".. ", string(s))
	var start, dst int

	i := 0

	mustclean := false
white:
	for i < len(s) {

		for ; i < len(s); i++ {
			if !asciiSpace[s[i]] {
				start = i
				break
			}
		}

		for ; i < len(s); i++ {
			mustclean = true
			if asciiSpace[s[i]] { // end of block
				ltmp := i - start // this is the correct length
				copy(s[dst:dst+ltmp], s[start:i+1])
				dst = dst + ltmp
				mustclean = false
				continue white
			}
		}
	}
	if mustclean { // hit end of input, so spit out leftovers
		ltmp := len(s) - start
		s = append(s[:dst], s[start:start+ltmp]...)
	} else { // If we finished in white space, we have to truncate
		s = s[:dst]
	}
	*sIn = s
	fmt.Printf("out: \"%s\"\n", string(s))
}

// RemoveByFields does the same, but using a library call
func removeByFields(sIn *[]byte) {
	s := *sIn
	s = bytes.Join(bytes.Fields(s), []byte(""))

	*sIn = s
}
