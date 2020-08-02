// Generate random sequences. At the moment, this is for testing, but they
// do not really belong in the test file.
// We do not work with "Seq" structures. We just make byte slices filled
// with characters.
// If one wants a Seq interface, that should be put in the seq files.

package randseq

import (
	"errors"
	"math/rand"
	"github.com/andrew-torda/goutil/seq"
)

// get_alfbt returns a pointer to the appropriate alphabet
func get_alfbt(typ seq.Seq_type) (alfbt []byte) {
	var dna_alfbt = []byte{'A', 'C', 'G', 'T'}
	var protein_alfbt = []byte{'a', 'c', 'd', 'e', 'f', 'g',
		'h', 'i', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'y'}
	switch typ {
	case seq.DNA:
		return dna_alfbt
	case seq.Protein:
		return protein_alfbt
	default:
		panic("program bug: unknown alphabet type")
	}
}

// randseq does not work with a sequence Seq. It works with a slice of
// bytes. We do use some definitions from the seq package, mainly
// to decide on how big the alphabet is.
func New(typ seq.Seq_type, n int) (s []byte) {
	alfbt := get_alfbt(typ)
	t := make([]byte, n)
	for i := range t {
		t[i] = alfbt[rand.Intn(len(alfbt))]
	}
	return t
}

// mutate changes some characters in s randomly.
// typ is either Protein or DNA. Rate is the probability of
// a change.
// The change happens in place, so you probably want to
// act on a copy.
// Return the number of positions that were actually changed.
func Mutate(typ seq.Seq_type, rate float32, s []byte) (n int) {
	alfbt := get_alfbt(typ)
	for i, c_old := range s {
		if rand.Float32() < rate {
			n++
			a := c_old
			for ; a == c_old; a = alfbt[rand.Intn(len(alfbt))] {
			}
			s[i] = a
		}
	}
	return
}

// del_rand randomly deletes some characters from a sequence
// It works in place.
// Return the number of characters deleted
func DelRand(rate float32, s []byte) (t []byte) {
	delme := make([]bool, len(s))
	for i := range s {
		if rand.Float32() < rate {
			delme[i] = true
		}
	}
	k := 0
	t = s
	for i := range s {
		if !delme[i] {
			t[k] = s[i]
			k++
		}
	}
	t = t[:k]
	return t
}

// Delete n elements from the byte slice.
// The deletion happens in place.
func DelN(n_to_del int, s []byte) (t []byte, err error) {
	if n_to_del < 0 {
		return nil, errors.New("DelN given negative number of places to delete")
	}
	if n_to_del >= len(s) {
		return nil, errors.New("randseq: n to delete as big as original slice")
	}

	delme := make([]bool, len(s))
	for i := 0; i < n_to_del; {
		for {
			n := rand.Intn(len(s))
			if delme[n] == false {
				delme[n], i = true, i+1
				break
			}
		}
	}
	k := 0
	t = s
	for i := range s {
		if !delme[i] {
			t[k] = s[i]
			k++
		}
	}
	t = t[:k]
	return t, err
}

// insertOne puts a single byte into a byte slice. We use the library
// function, so it will grow the slice as necessary.
// We can insert at the start or end, so the convention is, we insert
// before the index. That means, if we have a string of length n, inserting
// at 0 will put it at the start. Inserting at n will append it.
// In principle, it trashes the original slice.
func insertOne (pos int, s []byte, b byte) []byte {
	if pos > len(s) {
		pos = len(s)
	}
	second := make ([]byte, len(s) - pos)
	copy (second, s[pos:])
	s = s[:pos]
	s = append (s, b)
	s = (append (s, second...))
	return s
}

// InsN inserts n random characters
func InsN (typ seq.Seq_type, n_to_ins int, s[]byte) []byte {
	alfbt := get_alfbt(typ)
	for i := 0; i < n_to_ins; i++ {
		b := alfbt[rand.Intn(len(alfbt))]
		pos := rand.Intn (len(s))
		s = insertOne (pos, s, b)
	}
	return s
}
