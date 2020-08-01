// 31 July 2020

package randseq

import (
	"fmt"
	"io"
	"math/rand"
	"sync"
	//	"github.com/andrew-torda/goutil/seq"
)

const (
	nPadWhite = 9 // For padding for adding whitespace to sequences
)

// getseq returns a byte slice with a random sequence in it
func getseq(seqlen int, rnd *rand.Rand) []byte {
	space := seqlen + (seqlen / nPadWhite) // about 10% rubbish white space
	ret := make([]byte, seqlen, space)
	l := int32(len(letters))
	for i := 0; i < seqlen; i++ {
		ret[i] = letters[rand.Int31n(l)]
	}
	return ret
}

// RandSeqArgs is the set of arguments passed to the main function
type RandSeqArgs struct {
	Iseed int64     // random number seed
	Wrtr  io.Writer // where we write to
	Cmmt  string    // Comment for the sequences
	Nseq  int       // number of sequences
	Len   int       // Length of sequences
	NoGap bool      // Do not add gaps
	MkErr bool      // Add an error, by changing a length
}

var letters []byte

// addinner is used by addspace to add a space or newline
func addInner (s []byte, n int, c byte, spacernd *rand.Rand) []byte {
	for i := 0; i < n; i++ {
		s = append (s, 0)
		pos := rand.Int31n(int32(len(s)))
		copy (s[pos+1:], s[pos:])
		s[pos] = c
	}
	return s
}
	

// addspace is given a byte array and adds white characters at random
// positions. We work out how much space is to be used. We flip a coin.
// Heads we don't add a newline. Tails we make about 1/10 (integer 1/9)
// of the spaces to be newlines.
func addspace (s []byte, spacernd *rand.Rand) []byte {
	toAdd := cap(s) - len(s)
	coin := rand.Int31n(2)
	nNL := 0  // Number of new lines to add
	if coin == 0 {
		nNL = toAdd / 9
	}

	nSpace := toAdd - nNL
	s = addInner (s, nSpace, ' ', spacernd)
	s = addInner (s, nNL, '\n', spacernd)
	return s
}

// writeseq takes a bytestring which is our sequence. It adds a comment
// and sends it out for writing. n is the number of the sequence, so the
// output has comment lines "> something 1, > something 2..."
func writeseq(sChan <-chan []byte, args *RandSeqArgs, wg *sync.WaitGroup) {
	defer wg.Done()

	width := len (fmt.Sprintf ("%d", args.Nseq))
	spacernd := rand.New(rand.NewSource(args.Iseed))
	var i int
	for s := range sChan {
		i++
		s = addspace (s, spacernd)
		tmp := fmt.Sprintf ("> %s %[2]*d\n", args.Cmmt, width, i)
		args.Wrtr.Write([]byte(tmp))
		args.Wrtr.Write(s)
		args.Wrtr.Write([]byte{'\n'})
	}
}
// RandSeqMain writes random sequences to an io.Writer.
func RandSeqMain(args *RandSeqArgs) error {
	var wg sync.WaitGroup
	letters = []byte{'a', 'c', 'd', 'e', 'f', 'g',
		'h', 'i', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'y'}
	if args.NoGap == false {
		letters = append(letters, letters...)
		letters = append(letters, letters...)
		letters = append(letters, '-')
	}
	rnd := rand.New(rand.NewSource(args.Iseed))
	sChan := make(chan []byte)
	wg.Add(1)
	go writeseq(sChan, args, &wg)
	for i := 0; i < args.Nseq; i++ {
		s := getseq(args.Len, rnd)
		sChan <- s
	}
	close(sChan)
	wg.Wait()
	return nil
}
