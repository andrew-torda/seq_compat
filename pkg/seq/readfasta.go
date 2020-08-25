// Reader for fasta format files.

package seq

import (
	"bytes"
	"errors"
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
	cmmt       string // partial comment
	seq        []byte // partial string
	err        error  // error passed back to caller at end
	seqblock   []byte // Big block where all the sequences are placed
	term       byte   // terminator of comments or sequences
	diffLenSeq bool   // are all sequences the same length ?
	notfirst   bool   // Not the first call
}

const defaultReadSize = 2 * 1024

var rdsize int = defaultReadSize

// setFastaRdSize is only used during benchmarking to see the effect of
// buffer size.
func setFastaRdSize(i int) {
	if i <= 2 {
		panic("setFastaRdSize given buffer length of 2 or less")
	}
	rdsize = i
}

// NewItem is used by sync.pool.
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
						l.err = err // Real error (not EOF) occurred.
					}
					item.data = []byte("")
					item.complete = true
					l.ichan <- item // we have to flush
					close(l.ichan)
					return
				} else { //                Partial read. EOF, not an error
					l.input[n] = l.term // Add terminator
				}
			}
		}

		if ndx := bytes.IndexByte(l.input, l.term); ndx == -1 {
			item.data = l.input // No terminator found. Just send
			l.input = nil       // back whatever we have in the buffer.
			item.complete = false
		} else { //                         We did find a terminator
			item.data = l.input[:ndx]
			item.complete = true
			l.input = l.input[ndx+1:] // Set up for next loop
			if l.term == NL {
				l.term = cmmtChar
			} else {
				l.term = NL
			}
		}
		l.ichan <- item
	}
}

// firstCall is called when we have read up the first sequence and can
// allocate all the space we need.
func firstCall(l *lexer) {
	nseq, err := numseq.ByReading(l.rdr) // Swap to fancier version with
	if err != nil {                      // better memory allocation
		l.err = err
		return
	}
	if nseq < 1 {
		l.err = errors.New("no sequences found")
		return
	}
	lenseq := len(l.seq)
	l.seqblock = make([]byte, 0, lenseq*nseq)
}

type stateFn func(*lexer) stateFn

// seqFn is used to build up a sequence (not comment) and store it when complete.
// On the first call, we check if the sequences should be of the same length.
// If so, we allocate a single large block for sequences.
func seqFn(l *lexer) stateFn {
	item := <-l.ichan

	if item == nil || l.err != nil {
		return nil
	}

	white.Remove(&item.data)
	l.seq = append(l.seq, item.data...)
	complete := item.complete
	l.itempool.Put(item)
	if complete {
		if len(l.seq) == 0 {
			l.err = errors.New("Zero length sequence after" + l.cmmt)
			return nil
		}
		if !l.notfirst && !l.diffLenSeq { // on first sequence
			var tmp []byte
			l.notfirst = true
			firstCall(l)
			tmp = append(tmp, l.seq...)
			l.seq = l.seqblock
			l.seq = append(l.seq, tmp...)
		}
		seq := seq{cmmt: l.cmmt, seq : l.seq}
		l.seqgrp.seqs = append(l.seqgrp.seqs, seq)
		l.cmmt = ""
		if l.seqblock == nil { // Not using single block for sequences
			l.seq = nil
		} else {
			nlen := len(l.seqblock)
			l.seqblock = l.seqblock[:nlen+len(l.seq)]
			l.seq = l.seqblock[len(l.seqblock):]
		}
		return cmmtFn
	}
	return seqFn
}

// cmmtFn is used to build a function or save it when complete.
func cmmtFn(l *lexer) stateFn {
	item := <-l.ichan
	defer l.itempool.Put(item)
	if item == nil || l.err != nil {
		return nil
	}

	l.cmmt = l.cmmt + string(item.data)
	if item.complete {
		item.complete = false
		return seqFn
	}
	return cmmtFn
}

// ReadFasta reads fasta formatted files.
func ReadFasta(rdr io.ReadSeeker, seqgrp *SeqGrp, s_opts *Options) (err error) {
	l := lexer{
		rdr: rdr, ichan: make(chan *item), seqgrp: seqgrp, term: NL,
		diffLenSeq: s_opts.DiffLenSeq,
	}

	go l.next()
	for state := cmmtFn; state != nil; {
		state = state(&l)
	}
	if seqgrp.GetNSeq() == 0 {
		l.err = errors.New("No sequences found")
	}
	return l.err
}
