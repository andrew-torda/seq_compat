// 20 Dec 2017

// Package seq provides functions for sequences,
// which usually begin their lives in fasta format. It can
// read and write them.
//

// Organisation. Move everything to do with the seq structure to the start.
// Then think about moving all the seqgrp stuff to its own file.
// Big change I should try. At the moment, we allocate every sequence
// individually. I could allocate a big lump and set up pointers in there.
// Even more fun... Use golang.org/x/exp/mmap and just set up slices so
// they point in there.
package seq

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/andrew-torda/matrix"
	. "github.com/andrew-torda/seq_compat/pkg/seq/common"
)

// seq is the exported type.
type seq struct {
	cmmt string
	seq  []byte
}

// A marker to say what type of sequence we have, protein, DNA, ...
type SeqType byte

const (
	Unchecked SeqType = iota // Has not been looked at yet
	Unknown                  // Really unknown, not a protein or nucleotide
	Protein                  //
	DNA                      //
	RNA                      //
	Ntide                    // Nucleotide
)

// We only read ascii characters, so anything bigger than this is not
// valid.
const (
	MaxSym uint8 = 127
)

// Options contains all the choices passed in from the caller.
type Options struct {
	Vbsty        int
	ExpecttSeq   int  // Expected number of sequences
	DiffLenSeq   bool // false, unless we expect sequences to be different lengths
	Dry_run      bool // Do not write any files
	Keep_gaps_rd bool // Keep gaps upon reading
	Rmv_gaps_wrt bool // Remove gaps on output
}

// Constants
const cmmt_char byte = '>' // and this introduces comments in fasta format

// SeqGrp is a group of sequences, with some additional information
// such as what type (protein, nucleotide) and the number of symbols
// that have been used.
type SeqGrp struct {
	symUsed  [MaxSym]bool  // which symbols are actually used
	mapping  [MaxSym]uint8 // mapping['C'] tells me the index used for C
	revmap   []uint8       // revmap[2] tells me the character in place 2
	seqs     []seq
	counts   *matrix.FMatrix2d
	gapcnt   []int32 // count of gaps at each position
	stype    SeqType
	usedKnwn bool // Do we know how many symbols are used ?
	freqKnwn bool // are counts of symbols converted to fractional probabilities ?
}

// Function GetSeq returns the sequence as the original byte slice
func (s seq) GetSeq() []byte { return s.seq }

// Function GetCmmt returns the comment, including the leading ">"
func (s seq) GetCmmt() string { return s.cmmt }

// Function Len
func (s seq) Len() int { return len(s.seq) }

// SetSeq will replace whatever was the sequence with a new one
func (s *seq) SetSeq(t []byte) { s.seq = t }

// Clear gets rid of the contents of a sequence. If you want
// to delete a sequence, but it is part of an array, you can just
// clear its contents.
func (s *seq) Clear() {
	s.cmmt = ""
	s.seq = nil
}

// Empty returns true if a sequence has been cleared.
// We just check the sequence element of the structure.
func (s seq) Empty() bool {
	if len(s.seq) == 0 {
		return true
	}
	return false
}

// Gene_id returns the gene identifier for a sequence.
// Of course it does not really do that. It just returns the first
// word in the comment which is likely to be the gene identifier.
func (s seq) Gene_id() (gene_id string) {
	tmp := strings.Fields(s.cmmt)
	return tmp[0][:]
}

// Species tries to return the organism from which a sequence
// comes. Actually, it just looks in the comment line for a string
// between square brackets and returns it. Given
//     > xyz.123 comment here [  homo sapiens]
// it should return "homo sapiens" with leading and trailing white
// space removed.
func (s seq) Species() (species string, ok bool) {
	var i, j int
	if i = strings.LastIndexByte(s.cmmt, '['); i == -1 {
		return
	}
	if j = strings.LastIndexByte(s.cmmt, ']'); j == -1 {
		return
	}
	if i >= j { // Is this an error ?
		return
	} // We treat it as if there is no comment

	return strings.TrimSpace(s.cmmt[i+1 : j]), true
}

// Lower will change a sequence to lower case
// It is much smaller than the library version, since it only knows
// about characters that can occur in biological sequences.
// It also acts in place.
func (s *seq) Lower() {
	low := [256]byte{
		'A': 'a', 'B': 'b', 'C': 'c', 'D': 'd', 'E': 'e', 'F': 'f', 'G': 'g', 'H': 'h',
		'I': 'i', 'J': 'j', 'K': 'k', 'L': 'l', 'M': 'm', 'N': 'n', 'O': 'o', 'P': 'p',
		'Q': 'q', 'R': 'r', 'S': 's', 'T': 't', 'U': 'u', 'V': 'v', 'W': 'w', 'X': 'x',
		'Y': 'y', 'Z': 'z'}
	for i, c := range s.seq {
		if low[c] != 0 {
			s.seq[i] = low[c]
		}
	}
}

// trimBytes trims a slice to n bytes if it is longer
func trimStr(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// Upper changes a sequence to upper case, in place.
// It only works with bytes, not runes.
// It can return an error if it encounters a symbol it does
// not like (value higher than 128).
func (seq *seq) Upper() error {
	const diff = 'a' - 'A'
	const symerr = "bad sym \"%c\" at position %d starting \"%s\""
	s := seq.GetSeq()
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= MaxSym {
			t := seq.GetCmmt()
			t = trimStr(t, 40)
			e := fmt.Errorf(symerr, c, i, t)
			return e
		}
		if 'a' <= c && c <= 'z' {
			s[i] -= diff
		}
	}
	return nil
}

// Copy
func (s *seq) Copy() (t seq) {
	t = *new(seq)
	t.cmmt = s.cmmt
	t.SetSeq(s.GetSeq())
	return t
}

// String returns a sequence, with its comment at the start as
// a single string
func (s seq) String() (t string) {
	if len(s.cmmt) > 0 {
		t = fmt.Sprintf("%c%s\n", cmmt_char, s.GetCmmt())
	} else {
		t = ">\n"
	}
	t += string(s.GetSeq())
	return
}

// GetLen returns the length of the first sequence.
// If we are reading a multiple sequence alignment, this should be the length
// of all sequences.
func (seqgrp *SeqGrp) GetLen() int { return len(seqgrp.seqs[0].GetSeq()) }

// GetCounts gives us the normally non-exported counts
func (seqgrp *SeqGrp) GetCounts() *matrix.FMatrix2d {
	if seqgrp.counts == nil {
		seqgrp.UsageSite()
	}
	return seqgrp.counts
}

// GetSymUsed returns the normally non-exported symUsed
func (seqgrp *SeqGrp) GetSymUsed() [MaxSym]bool { return seqgrp.symUsed }

// TypeKnwn tells us if we have decided what kind of sequence we have.
func (seqgrp *SeqGrp) TypeKnwn() bool {
	if seqgrp.stype == Unchecked {
		return false
	}
	return true
}

// GetRevmap returns the non-exported revmap
func (seqgrp *SeqGrp) GetRevmap() []uint8 { return seqgrp.revmap }

// GetMapping returns the mapping (row) for a specific character
func (seqgrp *SeqGrp) GetMapping(c uint8) uint8 { return seqgrp.mapping[c] }

// clear gets rid of some calculated quantities. Useful for testing, but
// rarely for normal use. It is only exported in testing.
func (seqgrp *SeqGrp) clear() {
	for i := range seqgrp.symUsed {
		seqgrp.symUsed[i] = false
		seqgrp.mapping[i] = 255 // Any old silly number
	}
	seqgrp.revmap = nil
	seqgrp.counts = nil
	seqgrp.gapcnt = nil
	seqgrp.stype = Unchecked
	seqgrp.usedKnwn = false
	seqgrp.freqKnwn = false
}

// GetNSeq returns the number of sequences
func (seqgrp *SeqGrp) GetNSeq() int { return len(seqgrp.seqs) }

// GetNSym returns the number of symbols used in a seqgrp.
// Used in testing.
func (seqgrp *SeqGrp) GetNSym() int {
	if !seqgrp.usedKnwn {
		seqgrp.UsageSite()
	}
	if len(seqgrp.revmap) == 0 {
		seqgrp.mapsyms()
	}

	return len(seqgrp.revmap)
}

// GetSeqSlc return the slice of sequences
func (seqgrp *SeqGrp) GetSeqSlc() []seq { return seqgrp.seqs }

// GetMap tells us where we are storing info about a symbol in our
// tallies. So, seq[i].GetMap() tells us where to put info about this
// character.
func (seqgrp *SeqGrp) GetMap(c byte) uint8 { return seqgrp.mapping[c] }

// Upper uppercases all the members of a group of sequences.
func (seqgrp SeqGrp) Upper() error {
	for _, ss := range seqgrp.seqs {
		if err := ss.Upper(); err != nil {
			return err
		}
	}
	return nil
}

// check_seq_lengths should only be called if we are keeping
// gaps. Then we imagine all the sequences are aligned, so they
// must be the same length.
// For consistency, this should be callable on a seqgrp, not
// a slice of sequences.
func check_lengths(seq_set []seq) error {
	msg := `Sequence lengths are not the same. First sequence length %d, but
sequence %i length: %i. Sequence starts %s"`
	iwant := len(seq_set[0].GetSeq())
	for i := 1; i < len(seq_set); i++ {
		ilen := len(seq_set[i].GetSeq())
		if ilen != iwant {
			return (fmt.Errorf(msg, iwant, ilen, trimStr(seq_set[i].GetCmmt(), 40)))
		}
	}
	return nil
}

// Readfile takes a filename and reads sequences from it.
// each in turn. It returns a SeqGrp and error.
func Readfile(fname string, s_opts *Options) (*SeqGrp, error) {
	var seqgrp = new (SeqGrp)
	var err error
	var fp io.ReadCloser // don't use a file. It could be stdin.

	if fname != "" {
		if fp, err = os.Open(fname); err != nil {
			return nil, err
		}
	} else {
		fp = os.Stdin
	}

	defer fp.Close()

	if err := ReadFasta(fp, seqgrp, s_opts); err != nil {
		return seqgrp, err
	}

	if s_opts.Keep_gaps_rd {
		check_lengths(seqgrp.seqs)
	}
	return seqgrp, err
}

// WriteToF takes a filename and a slice of sequences.
// It writes the sequences to the file.
// For each sequence, it should check if the sequence has been
// set to nil.
// What I could change: If we are removing gaps, we make a buffer which grows
// character by character via WriteByte(). I could make a buffer beforehand
// and grow as necessary.
// This should also really act on a seqgrp.
func WriteToF(outseq_fname string, seq_set []seq, s_opts *Options) (err error) {
	const c_per_line = 60
	var nilstring string
	var outfile_fp io.Writer
	switch {
	case s_opts.Dry_run:
		outfile_fp = ioutil.Discard
	case outseq_fname == nilstring:
		outfile_fp = os.Stdout
	default:
		t, err := os.Create(outseq_fname)
		if err != nil {
			return fmt.Errorf("Creating output sequence file: %v", err)
		}
		defer t.Close()
		outfile_fp = t
	}

	var t []byte
	for _, seq := range seq_set {
		if seq.Empty() {
			continue
		}
		fmt.Fprintf(outfile_fp, "%c%s\n", cmmt_char, seq.GetCmmt())

		s := seq.GetSeq()
		if s_opts.Rmv_gaps_wrt { // we have to remove gap characters on output
			n := 0
			for i := range s { //    So we start by looking how many non-gap
				if s[i] != GapChar { //  characters there are.
					n++
				}
			}
			if cap(t) < n { // See if our scratch space is big enough
				t = make([]byte, n)
			}

			m := 0
			for i := range s {
				if s[i] != GapChar {
					t[m] = s[i]
					m++
				}
			}
			s = t[:n]
		}
		for ; len(s) > c_per_line; s = s[c_per_line:] {
			fmt.Fprint(outfile_fp, string(s[:c_per_line]), "\n")
		}
		fmt.Fprint(outfile_fp, string(s), "\n")
	}
	return
}

// FindNdx Returns the index of the sequence containing a string.
// Numbering starts from zero. We remove any ">", space or tab at the start.
func (seqgrp *SeqGrp) FindNdx(s string) int {
	s = strings.TrimLeft(s, " >	")

	for i, seq := range seqgrp.seqs {
		if strings.Contains(seq.GetCmmt(), s) {
			return i
		}
	}
	return -1
}

// Str2SeqGrp takes some strings and returns them as a seqgrp.
// sIn is a slice of strings which are the sequences.
// prefix is an optional argument. Sequences need names/comments. If
// prefix is not given, sequences will be called "> s1", "> s2", ...
func Str2SeqGrp(sIn []string, prefix ...string) (*SeqGrp) {
	var base string
	seqgrp := new (SeqGrp)
	if prefix == nil {
		base = "s"
	} else {
		base = prefix[0]
	}
	for i, s := range sIn {
		f := seq{cmmt: fmt.Sprint(base, i), seq: []byte(s)}
		seqgrp.seqs = append(seqgrp.seqs, f)
	}
	return seqgrp
}
