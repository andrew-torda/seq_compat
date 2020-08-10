// Reader for fasta format files.

package seq

import (
	"bytes"
	"io"
	"sync"

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
	input    []byte
	ichan    chan *item
	seqgrp   *SeqGrp
	rdr      io.Reader
	itempool sync.Pool
	cmmt     string // partial comment
	seq      []byte // partial string
	term     byte
	err      error
}

const defaultReadSize = 512

var rdsize int = defaultReadSize

// setFastaRdSize is only used during benchmarking
func setFastaRdSize(i int) { rdsize = i }

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
					close(l.ichan)
					return
				} else { // Partial read. EOF, not an error
					l.input[n] = cmmtChar // Add terminator
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

type stateFn func(*lexer) stateFn

func gstart(l *lexer) stateFn {
	return gstart
}

// This is a function which might use a bit of memory. Replace it later
func removeWhite(s []byte) []byte {
	f := bytes.Fields(s)
	r := bytes.Join(f, []byte(""))
	return r
}

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
		seq := seq{cmmt: l.cmmt, seq: l.seq}
		l.seqgrp.seqs = append(l.seqgrp.seqs, seq)
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

// Readfasta is my second version of a fasta reader.
func ReadFasta(rdr io.Reader, seqgrp *SeqGrp, s_opts *Options) (err error) {
	l := lexer{rdr: rdr, ichan: make(chan *item, 2), seqgrp: seqgrp, term: NL}

	go l.next()
	for state := gcmmt; state != nil; {
		state = state(&l)
	}
	return nil
}
