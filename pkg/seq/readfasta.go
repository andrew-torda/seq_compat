// Reader for fasta format files.
// Memory explanation...
// There are three ways to allocate memory, crammed into one function.
//  1. If we have sequences of different lengths, l.seq is cleared after
//     each sequence and we just keep appending to it. l.seq becomes part
//     of the new sequence. This means each sequence is its own allocation.
//     We could change this to put all sequences into one block which grows.
//  2. If we have a set of aligned sequences, they are all the same length
//     and we probably are keeping gaps. We do a scan of the file, estimate
//     the number of sequences. We read the first sequence to find the length.
//     Then allocate one block for all sequences. We save the first sequence
//     and set l.seq to point directory into this block. There should be no
//     more unnecessary copying of bytes from sequences.
//  3. Sequences are the same length, but we only want to keep a range from
//     each sequence. We allocate a block for all sequences as in case 2.
//     l.seq is a re-used buffer. After each sequence is complete, we copy
//     the region we want into l.seqblock and set the used part of the buffer
//     to length zero.

package seq

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/andrew-torda/seq_compat/pkg/numseq"
	"github.com/andrew-torda/seq_compat/pkg/seq/common"
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
	seqblock   []byte // Big block where all the sequences are placed
	err        error  // error passed back to caller at end
	expLen     int    // Expected length of sequences. If zero, not used.
	rangeStart int    // Start and end of sequence range to be kept. Copied
	rangeEnd   int    // from seq options. Zero means keep everything.
	term       byte   // terminator of comments or sequences
	memtype    byte   // diff length sequences, same or a range from each seq
	RmvGapsRd  bool   // copied from s_opts
	ZeroLenOK  bool   // from s_opts, zero length seqs OK
	notfirst   bool   // Not the first call
}

const defaultReadSize = 4 * 1024

var rdsize int = defaultReadSize

// setFastaRdSize is only used during benchmarking to see the effect of
// buffer size.
func setFastaRdSize(i int) {
	if i <= 2 {
		panic("setFastaRdSize given buffer length of 2 or less")
	}
	rdsize = i
}

const (
	diffLen byte = iota
	sameLen
	withRange
)

// NewItem is used by sync.pool.
func newItem() interface{} { return new(item) }

// next reads from the input and sends an item to channel, ichan.
// An item is terminated by l.term, or the end of the buffer or
// end of input.
// Use a pair of buffers for reading. When one is being filled, the other might
// be processed by the comment or sequence reading function.
func (l *lexer) next() {
	l.itempool.New = newItem
	backbuf1 := make([]byte, rdsize)
	backbuf2 := make([]byte, rdsize)
	var first bool = true
	curbuf := &backbuf2
	for {
		item := l.itempool.Get().(*item)
		if len(l.input) == 0 {
			if curbuf == &backbuf1 {
				curbuf = &backbuf2
			} else {
				curbuf = &backbuf1
			}
			l.input = (*curbuf)[:]
			if n, err := l.rdr.Read(l.input); n != rdsize { // EOF or error?
				l.input = l.input[:n]
				if n == 0 { // really finished
					if err != nil && err != io.EOF {
						l.err = err // Real error (not EOF) occurred.
					}
					item.data = nil
					item.complete = true
					l.ichan <- item // we have to flush
					close(l.ichan)
					return
				} else { // Partial read. EOF, not an error
					if l.input[n-1] != l.term {
						l.input = append(l.input, l.term) // Add terminator
					}
				}
			}
		}
		if first { // skip over leading ">"
			first = false
			l.input = l.input[1:]
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
func firstCall(l *lexer) error {
	const invalidRange = "invalid seq range %d to %d, length is only %d"
	nseq, err := numseq.ByReading(l.rdr)
	if err != nil {
		return err
	}
	if nseq < 1 {
		return errors.New("no sequences found")
	}
	l.expLen = len(l.seq)
	var sz int
	if l.rangeEnd >= l.expLen {
		return fmt.Errorf(invalidRange, l.rangeStart, l.rangeEnd, l.expLen)
	}
	if l.rangeStart != 0 || l.rangeEnd != 0 {
		sz = l.rangeEnd - l.rangeStart + 1
	} else {
		sz = l.expLen
	}
	l.seqblock = make([]byte, 0, sz*nseq)
	return nil
}

type stateFn func(*lexer) stateFn

// seqFn is used to build up a sequence (not comment) and store it when complete.
// On the first call, we check if the sequences should be of the same length.
// If so, we allocate a single large block for sequences.
func seqFn(l *lexer) stateFn {
	const bustLen = "seqs not same length, wanted %d, got %d"
	item := <-l.ichan
	if item == nil || l.err != nil {
		return nil
	}

	white.Remove(&item.data)
	if l.RmvGapsRd {
		white.CharRemove(&item.data, common.GapChar)
	}
	l.seq = append(l.seq, item.data...)
	complete := item.complete
	l.itempool.Put(item)
	if complete {
		if len(l.seq) == 0 {
			if !l.ZeroLenOK { // zero length seqs usually not OK
				l.err = errors.New("Zero length sequence after" + l.cmmt)
				return nil
			}
		}

		if !l.notfirst && (l.memtype == sameLen || l.memtype == withRange) {
			var tmp []byte
			l.notfirst = true
			if l.err = firstCall(l); l.err != nil { // Just allocates the space
				return nil
			}
			if l.memtype == sameLen {
				tmp := append(tmp, l.seq...)  // Save first seq
				l.seq = l.seqblock            // where we will be saving all seqs
				l.seq = append(l.seq, tmp...) // save slice we put aside
			}
		}
		if l.memtype == sameLen || l.memtype == withRange {
			if l.expLen != len(l.seq) {
				l.err = fmt.Errorf(bustLen, l.expLen, len(l.seq))
				return nil
			}
		}

		var vseq seq
		switch l.memtype {
		case diffLen:
			vseq = seq{cmmt: l.cmmt, seq: l.seq}
		case sameLen:
			vseq = seq{cmmt: l.cmmt, seq: l.seq}
		case withRange:
			toUse := l.seq[l.rangeStart : l.rangeEnd+1]
			start := len(l.seqblock)
			l.seqblock = append(l.seqblock, toUse...)
			vseq = seq{cmmt: l.cmmt, seq: l.seqblock[start : start+len(toUse)]}
		}

		l.seqgrp.seqs = append(l.seqgrp.seqs, vseq)
		l.cmmt = ""
		switch l.memtype {
		case diffLen:
			l.seq = nil // Not using single block for sequences. Forces fresh memory on next.
		case sameLen:
			nlen := len(l.seqblock)
			l.seqblock = l.seqblock[:nlen+len(l.seq)]
			l.seq = l.seqblock[len(l.seqblock):]
		case withRange:
			l.seqblock = append(l.seqblock, l.seq[l.rangeStart:l.rangeEnd]...)
			l.seq = l.seq[:0]
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

// memtype decides how we will store the sequences. If they are of different
// lengths, they just get individually allocated. If they are the same length,
// we first allocate a pool and adjust pointers within the pool. If we are
// only going to keep a piece of each sequence, we allocate the pool, but
// read sequences into a temporary buffer, but then copy over the
// wanted bits of the buffer.
func memtype(s_opts *Options) byte {
	if s_opts.DiffLenSeq {
		return diffLen
	}
	if s_opts.RangeStart == 0 && s_opts.RangeEnd == 0 {
		return sameLen
	}
	return withRange
}

// checkBroken looks for inconsistent flags or code problems
func checkBroken(s_opts *Options) error {
	if s_opts.DiffLenSeq && (s_opts.RangeStart != 0 || s_opts.RangeEnd != 0) {
		return errors.New("Can't use range option. Sequences of different lengths")
	}
	return nil
}

// ReadFasta reads fasta formatted files.
func ReadFasta(rdr io.ReadSeeker, seqgrp *SeqGrp, s_opts *Options) (err error) {
	if err := checkBroken(s_opts); err != nil {
		return err
	}
	l := lexer{
		rdr: rdr, ichan: make(chan *item), seqgrp: seqgrp, term: NL,
		RmvGapsRd:  s_opts.RmvGapsRd,
		rangeStart: s_opts.RangeStart, rangeEnd: s_opts.RangeEnd,
		ZeroLenOK: s_opts.ZeroLenOK,
		memtype:   memtype(s_opts),
	}

	go l.next()
	for state := cmmtFn; state != nil; {
		state = state(&l)
	}
	if l.err != nil {
		return l.err
	}
	if seqgrp.NSeq() == 0 {
		return errors.New("No sequences found")
	}
	return nil
}
