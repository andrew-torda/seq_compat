package seq

import (
	"bytes"
)

// Given two strings as slices, see if they overlap like this
//abcdef
//   defghi
//If the second string can be, append it to the first string.
// This happens in place, so the first string will be overwritten.
//

func shift_s1(t1, t2 []byte, i_want int) (ifound int, increased bool) {
	if i_want > len(t1) {
		return
	}
	tail := t1[len(t1)-i_want:]
	for i := 0; i != -1; i++ {
		i_ext := bytes.Index(t2[i:], tail)
		if i_ext == -1 { // no possibilities, just return
			return
		}
		i += i_ext

		candidate := t2[:i+len(tail)]
		var lc = len(candidate)
		var lt = len(t1)
		if lc > lt {
			return
		}
		tail := t1[len(t1)-len(candidate):]
		if bytes.Equal(tail, candidate) {
			return len(candidate), true
		}
	}
	return
}

func find_one(s1, s2 []byte, min_ovlp int) (ifound int) {
	increased := true
	var got int
	for ; increased == true && min_ovlp <= len(s1); min_ovlp = ifound + 1 {
		got, increased = shift_s1(s1, s2, min_ovlp)
		if got > ifound {
			ifound = got
		}
	}
	return
}

func find_ovlp(s1, s2 []byte, min_ovlp int) int {
	ifirst := find_one(s1, s2, min_ovlp)
	isecond := find_one(s2, s1, min_ovlp)
	if ifirst > isecond {
		return ifirst
	}
	return -(isecond)
}

// Given the overlap detected between two strings, merge them and
// return a new string.
// Now that I am swapping to a byte type, this should change and perhaps become
// a simple append operation
/*
func merge_str(s1, s2 []byte, ovlp int) (s_merge []byte) {
	switch {
	case ovlp > 0:
		s_merge = s1 + s2[ovlp:]
	case ovlp < 0:
		ovlp = -(ovlp)
		s_merge = s2 + s1[ovlp:]
	default:
		log.Fatal("merge_str() called with no overlap")
	}
	return s_merge
}
*/
// This is written deliberately not to use the ability to return multiple values.
// We return true / false, so the caller can use this in switch or if statements.
// If we find overlap, we return it via the int pointer
func (s Seq) ovlp_exists(t Seq, ovlp *int, min_ovlp int) bool {
	*ovlp = find_ovlp(s.seq, t.seq, min_ovlp)
	if *ovlp == 0 {
		return false
	}
	return true
}

// merge returns the merger of s and t according to the overlap given by ovlp.
func (s Seq) merge(t Seq, ovlp int) (r []byte) {
	switch {
	case ovlp > 0:
		r = append(s.seq, t.seq[ovlp:]...)
	case ovlp < 0:
		ovlp = -(ovlp)
		r = append(t.seq, s.seq[ovlp:]...)
	default:
		panic("program bug: overlap called with zero value")
	}
	return
}
