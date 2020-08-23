// Reader for fasta format files.

package seq

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/andrew-torda/seq_compat/pkg/numseq"
	"github.com/andrew-torda/seq_compat/pkg/white"
)

// An item is terminated by a newline if we are in a comment or a comment
// character ">" if we are in a sequence.
const (
	NL       = '\n'
	cmmtChar = '>'
)

type item struct {
	data     []byte
	complete bool
}

type lexer struct {
	input      []byte
	ichan      chan *item
	seqgrp     *SeqGrp
	rdr        io.ReadSeeker
	itempool   sync.Pool
	cmmt       string   // partial comment
	seq        []byte   // partial string
	term       byte     // terminator of comments or sequences
	err        error    // error passed back to caller at end
	DiffLenSeq bool     // are all sequences the same length
	fAddSeq    addSeqFn // function to be used for adding sequences
	seqblock   []byte   // Big block where all the sequences are placed
}

const defaultReadSize = 512

var rdsize int = defaultReadSize

// setFastaRdSize is only used during benchmarking
func setFastaRdSize(i int) {
	if i <= 2 {
		panic("setFastaRdSize given buffer length of 2 or less")
	}
	rdsize = i
}

func newItem() interface{} { return new(item) }

// next reads from the input and sends an item to channel, ichan.
// An item is terminated by l.term, or the end of the buffer or
// end of input.
func (l *lexer) next() {
	l.itempool.New = newItem
	for {
		item := l.itempool.Get().(*item)
		if len(l.input) == 0 {
			l.input = make([]byte, rdsize)
			if n, err := l.rdr.Read(l.input); n != rdsize {
				if n == 0 {
					if err != nil && err != io.EOF {
						l.err = err // signal that a real error occurred.
					}
					item.data = []byte("")
					item.complete = true
					l.ichan <- item // we have to flush
					close(l.ichan)
					return
				} else { // Partial read. EOF, not an error
					l.input[n] = l.term // Add terminator
				}
			}
		}

		if ndx := bytes.IndexByte(l.input, l.term); ndx == -1 {
			item.data = l.input // no terminator found, so just send
			l.input = nil       // back whatever we have in the buffer.
			item.complete = false
		} else { //                                We did find a terminator
			newlInput := l.input[ndx+1:] //        Advance pointer
			item.data = l.input[:ndx]    //
			item.complete = true         //
			l.input = newlInput          //        Set up for next loop
			if l.term == NL {
				l.term = cmmtChar
			} else {
				l.term = NL
			}
		}
		l.ichan <- item
	}
}

// fAppndSeq takes a byte slice and appends it to the slice of byte slices that
// are the sequences
func fAppndSeq(l *lexer, s seq) {
	l.seqgrp.seqs = append(l.seqgrp.seqs, s)
}

// fAddInBlockSeq takes a byte slice, appends it to our big block and
// then sets the pointers in the slice of sequences
func fAddInBlockSeq(l *lexer, s seq) {
	l.seqblock = append(l.seqblock, s.seq...)
}

// faddseqfirst // is called when we have the first sequence ready.
// It then decides if future sequences should be added by fAppndSeq (simply
// appending them) or by fBlockSeq which first allocates a block of sequences.
func fAddSeqFirst(l *lexer, s seq) {
	fmt.Println("in faddseqfirst")
	if l.DiffLenSeq {
		l.fAddSeq = fAppndSeq
		fAppndSeq(l, s)
		return
	}
	//  We now get to do the fancier, memory saving allocation.
	fmt.Println("add function for subsequent sequences")
	nseq, err := numseq.ByReading(l.rdr)
	if err != nil {
		l.err = err
		return
	}
	lenseq := len(s.seq)
	l.seqblock = make([]byte, 0, lenseq*nseq)
	l.fAddSeq = fAddInBlockSeq
	fAddInBlockSeq(l, s)
}

type stateFn func(*lexer) stateFn

// We are reading a sequence
func gseq(l *lexer) stateFn {
	item := <-l.ichan
	defer l.itempool.Put(item)
	if item == nil || l.err != nil {
		return nil
	}

	white.Remove(&item.data)
	l.seq = append(l.seq, item.data...)

	if item.complete {
		if len(l.seq) == 0 {
			l.err = errors.New("Zero length sequence after" + l.cmmt)
		}
		seq := seq{cmmt: l.cmmt, seq: l.seq}
		//		l.seqgrp.seqs = append(l.seqgrp.seqs, seq)
		//		fAppndSeq(l, seq)
		l.fAddSeq(l, seq)
		l.cmmt = ""
		l.seq = nil
		return gcmmt
	}
	return gseq
}

// We are reading a comment
func gcmmt(l *lexer) stateFn {
	item := <-l.ichan
	defer l.itempool.Put(item)
	if item == nil || l.err != nil {
		return nil
	}

	l.cmmt = l.cmmt + string(item.data)
	if item.complete {
		item.complete = false
		return gseq
	}
	return gcmmt
}

// ReadFasta reads fasta formatted files.
func ReadFasta(rdr io.ReadSeeker, seqgrp *SeqGrp, s_opts *Options) (err error) {
	l := lexer{
		rdr: rdr, ichan: make(chan *item), seqgrp: seqgrp, term: NL,
		DiffLenSeq: s_opts.DiffLenSeq, fAddSeq: fAddSeqFirst,
	}

	go l.next()
	for state := gcmmt; state != nil; {
		state = state(&l)
	}
	if seqgrp.GetNSeq() == 0 {
		l.err = errors.New("No sequences found")
	}
	return l.err
}

// --------------- experimenting ----------------
type addSeqFn func(*lexer, seq) // Adds a sequence to a slice of sequences
type rdInfo struct {
	seqSpace []byte   // Maybe where we will allocate space for sequences.
	addFn    addSeqFn // The function for adding a new sequence
	sameLen  bool     // are all sequences the same length
}

// setupSeqAlloc sets up for the coming reading of sequences.
// A tiny trick. We want to store state for the duration of the reading.
// We could have some variables local to the file, but I would forget
// them and they hang around.
// If we stick them in the lexer structure, they will be cleaned up
// as soon as the lexer goes away.
// During programming, we can make it a file variable. Later, we can
// move it into the lexer structure.
/*
var rdInf rdInfo

func setupSeqAlloc(rdr io.ReadSeeker, s_opts *Options) error {
	if s_opts.DiffLenSeq { // Sequences have different lengths. Just use
		return nil //            default append operation
	}
	nseq, err := numseq.ByReading(rdr)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "setup alloc thinks there are", nseq, "sequences")
	return nil
}
*/
