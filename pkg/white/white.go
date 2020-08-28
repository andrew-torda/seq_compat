// Removing white space from bytes (ascii strings), but do it in place.
// It does not matter which of my loops one uses. They take the same time.
// Do not use the version that looks like
// func removeByFields(sIn *[]byte) {
//	s := *sIn
//	s = bytes.Join(bytes.Fields(s), []byte(""))
//
//	*sIn = s
//}
	
// It is short and cute, but is slower and uses a lot of memory.
	
package white

// Remove acts on a pointer to a byte slice, in place and removes all the
// white space.
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

// removeWithBlocks does the same as Remove and has the same signature.
// It is an alternative which might be faster if you have relatively little
// white space.
func removeWithBlocks(sIn *[]byte) {
	var asciiSpace = [256]bool{
		0: true, '\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true,
	}

	s := *sIn

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
				ltmp := i - start // lengt of block to copy
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
}
