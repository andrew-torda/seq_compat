// 20 Dec 2017

// Package seq provides functions for sequences, which usually begin their lives in fasta format. It can
// read write and compare them.
package seq

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
)

// Seq is the exported type.
type Seq struct {
	cmmt string
	seq  []byte
}


type Seq_type byte

// DNA or RNA or protein ?
const (
	DNA Seq_type = iota
	Protein
)

// Options contains all the choices passed in from the caller.
type Options struct {
	Vbsty      int
	Dsbl_merge bool // disable merging of sequences
	Dry_run    bool // Do not write any files
	Keep_gaps  bool // Keep gaps upon reading
	Min_ovlp   int  // minimimum overlap of sequences
	Rmv_gaps   bool // Remove gaps on output
}

// These const types are what the compare function can return when it
// compares two sequences.
const (
	No_sim = iota
	Identical
	Ident_remove_frst // identical if we say abcde is the same as ab--e
	Ident_remove_scnd // remove first or second, whoever has most gaps
	S_in_t            // first is substring of second
	T_in_s            // second is substring of first
	Ovlp_merged       // there is overlap and we will return a merged sequence
)

// Constants
const gapchar byte = '-'   // a minus sign is always used for gaps
const cmmt_char byte = '>' // and this introduces comments in fasta format

// Contains reports if s contains t. That is, is t a subsequence of s.
func (s Seq) Contains(t Seq) bool {
	return bytes.Contains(s.seq, t.seq)
}

// cmp_with_gaps takes two sequences and says if they are identical.
// They must have the same length, since they come from a gapped alignment.
// We count the number of gap characters and return an indicator of who
// had the most gaps. Presumably, this is the candidate to remove.
func cmp_with_gaps(s, t Seq) int {
	var n_s, n_t int
	for i := range s.seq { //      If characters are identical
		if s.seq[i] == t.seq[i] { // or if one of them is a gap,
			continue //              we just keep going.
		}
		if s.seq[i] == gapchar {
			n_s++
			continue
		}
		if t.seq[i] == gapchar {
			n_t++
			continue
		}
		return -1
	} // Sequences are identical in non-gaps. Decide which should go
	if n_s > n_t {
		return Ident_remove_frst
	}
	return Ident_remove_scnd
}

// Compare acts on a sequence s and compares it to sequence t.
// It does lots of checks and returns one of the constant values defined
// above.
func Compare(s, t Seq, s_opts *Options) (sim int, s_merged []byte) {
	if s.Empty() || t.Empty() {
		return No_sim, nil
	}
	s_len, t_len := s.size(), t.size()
	if s_len == t_len { //     Could be that they
		if bytes.Equal(s.seq, t.seq) { // really are identical or
			return Identical, nil
		} else if s_opts.Keep_gaps { // a gapped alignment makes them the same length
			differ := cmp_with_gaps(s, t)
			switch differ {
			case Ident_remove_frst:
				return Ident_remove_frst, nil
			case Ident_remove_scnd:
				return Ident_remove_scnd, nil
			}
		}
	}

	if s_len < t_len {
		if t.Contains(s) {
			return S_in_t, nil
		}
	} else {
		if s.Contains(t) {
			return T_in_s, nil
		}
	}
	if !s_opts.Dsbl_merge { // if merging  has not been disabled,
		var ovlp_siz int
		if s.ovlp_exists(t, &ovlp_siz, s_opts.Min_ovlp) {
			return Ovlp_merged, s.merge(t, ovlp_siz)
		}
	}
	return No_sim, nil
}

// seq.Get_cmmt returns the comment from a sequence
func (s Seq) get_cmmt() string {
	return s.cmmt[:]
}

// seq.get_seq returns the sequence from a sequence.
// So far, I have not had to export this.
func (s Seq) get_seq() []byte {
	return s.seq
}

// Fix me
func (s Seq) Get_seq ()  []byte {return s.get_seq()}

// Size returns the size of a sequence
func (s Seq) size() int {
	return len(s.seq)
}

func (s Seq) Testsize() int {
	return len(s.seq)
}

// Set_seq will replace whatever was the sequence with a new one
func (s *Seq) Set_seq(t []byte) {
	s.seq = t
}

// Clear gets rid of the contents of a sequence. If you want
// to delete a sequence, but it is part of an array, you can just
// clear its contents.
func (s *Seq) Clear() {
	s.cmmt = ""
	s.seq = nil
}

// Empty returns true if a sequence has been cleared.
// We just check the sequence element of the structure.
func (s Seq) Empty() bool {
	if len(s.seq) == 0 {
		return true
	}
	return false
}

// Gene_id returns the gene identifier for a sequence.
// Of course it does not really do that. It just returns the first
// word in the comment which is likely to be the gene identifier.
func (s Seq) Gene_id() (gene_id string) {
	tmp := strings.Fields(s.cmmt)
	return tmp[0][:]
}

// Species tries to return the organism from which a sequence
// comes. Actually, it just looks in the comment line for a string
// between square brackets and returns it. Given
//     > xyz.123 comment here [  homo sapiens]
// it should return "homo sapiens" with leading and trailing white
// space removed.
func (s Seq) Species() (species string, ok bool) {
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
// It also acts in place
func (s Seq) Lower() {
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

// myscanner is used when reading sequences.
// It it tied to a bufio.Reader, which is really a read-only file.
// b is where pieces of text are read into and then split into comment
// and sequence. It will not go away until the myscanner goes away.
// This avoid a boatload of allocations (one per sequence).
type myscanner struct {
	bufio_reader *bufio.Reader
	b            []byte
	eof          bool
	err          error
}

func newmyscanner(fp io.Reader) *myscanner {
	r := new(myscanner)
	r.bufio_reader = bufio.NewReader(fp)
	return r
}

// cleanup is not necessary, but I like it
func (scnr *myscanner) cleanup() {
	scnr.bufio_reader = nil
	scnr.b = nil
}

// with_nl is a necessary evil. usually, we search for ">" to get the start of the
// next sequence. It must be at the start of a line.
// It can happen that someone has buried a ">" in the middle of a line.
// Check for this. If it has happened, then we need to allocate a new buffer, append
// into it and return a slice with this buffer.
func (scnr *myscanner) with_nl(delim byte) (line []byte) {
	line = scnr.readbigslice(delim)
	if len(line) < 1 {
		return
	}
	if len(line) > 2 { // Normal common case
		l := len(line) - 2
		if line[l] == '\n' || scnr.eof == true {
			return line
		}
	}
	var a []byte                                    // The comment line has broken our parser
	var tmp_line []byte = make([]byte, len(line))   // Save what we have in a newly
	copy(tmp_line, line)                            // allocated buffer
	a, scnr.err = scnr.bufio_reader.ReadBytes('\n') // Just go to end of line
	if scnr.err != nil {
		return
	}

	tmp_line = append(tmp_line, a...) // Stick it on the end and keep going
	a = scnr.readbigslice(delim)      // This will go until the ">"
	line = append(tmp_line, a...)     // Now we do not need tmp_line any more
	return
}

// readbigslice is a wrapper around ReadSlice.
// If it is happy with a single read, it just returns it so we avoid
// allocating and copying as happens with ReadBytes().
// If the string is too big, then we have to allocate space.
func (scnr *myscanner) readbigslice(delim byte) (line []byte) {
	line, scnr.err = scnr.bufio_reader.ReadSlice(delim)

	if scnr.err == nil && len(line) > 0 { //  Most common case
		return //                             Just return
	}

	if scnr.err == io.EOF {
		scnr.eof = true
		scnr.err = nil
		return
	}

	if scnr.err == bufio.ErrBufferFull { // Now the nasty case of having to do multiple reads
		var r []byte
		r = append(r, line...)
		line = r
		var tmp []byte
		for scnr.err == bufio.ErrBufferFull {
			line = append(line, tmp...)
			tmp, scnr.err = scnr.bufio_reader.ReadSlice(delim)
		}
		line = append(line, tmp...)
	}
	return
}

// get_next_lump () returns the text corresponding to exactly one sequence,
// with its comment line and then the actual sequence.
// Return text up to the next ">" character or EOF.
// Return true if we are happy.
//
//
func (scnr *myscanner) get_next_lump() bool {
	scnr.b = scnr.with_nl(cmmt_char)
	if len(scnr.b) < 2 {
		if scnr.err == io.EOF {
			scnr.eof = true
			scnr.err = nil // Not an error, but tell the caller there is no more data
		}
		return false
	}
	if scnr.err == io.EOF { // do not signal an error
		scnr.err = nil // On the next call, we will return no more data.
		scnr.eof = true
	}
	return true
}

// lump_split takes a lump of characters which should contain a comment,
// followed by the sequence. The comment is delimited by a newline.
// The sequence can have any amount of white space in it.
//
func lump_split(b []byte, white []bool, scnr *myscanner) (seq Seq, err error) {
	if len(b) < 2 {
		err = errors.New("lump_split: too short")
		return
	}
	n := bytes.IndexByte(b, '\n')
	if n < 0 {
		err = errors.New("lump_split: no newline")
		return
	}
	seq.cmmt = string(bytes.TrimRightFunc(b[:n], unicode.IsSpace))
	b = b[n+1:] // First byte after the newline
	nw := 0
	for _, c := range b {
		if !white[c] { // Count the number of bytes we need (non-white)
			nw++
		}
	}
	seq.seq = make([]byte, nw)
	i := 0
	for _, c := range b {
		if !white[c] {
			seq.seq[i] = c
			i++
		}
	}
	return
}

// seq.ReadSeqs takes a filename as input and reads sequences in fasta
// format.
// It returns ndone and error.. The number of sequences read up and an error.
// It should work with utf-8 files.
// This should not be the case with sequences, but it might be the case with comments.
// The function will stop reading if it encounters an error.
func readSeqs(fname string, seq_set []Seq, seq_map map[string]int, s_opts *Options) (seq_out []Seq, n_dup int, err error) {
	fp, err := os.Open(fname)
	if err != nil {
		return
	}
	defer fp.Close()
	s := newmyscanner(fp)
	{ // Our scanner eats '>' characters, but our
		var btmp byte // first line has not been through scanner,
		if btmp, err = s.bufio_reader.ReadByte(); err != nil {
			return //            so we jump over first character.
		}
		if btmp != cmmt_char { // Since we are here, we can check the file format
			err = fmt.Errorf("First byte in file was not a comment character")
			return
		}
	}

	seq_out = seq_set
	white := [256]bool{cmmt_char: true, //     Set of characters we do not want
		'\t': true, '\n': true, '\v': true, // in sequences, including the
		'\f': true, '\r': true, ' ': true} //  comment char '>'

	if !s_opts.Keep_gaps { //      Unless we want to keep gaps, we also
		white[gapchar] = true //   remove "-" characters. Treat them as white space
	}

	const dup_warn = "Sequence starting \"%s\" was duplicated in file %s\n"

	nc := 0

	for s.get_next_lump() {
		nc++
		const h_len = 40 // We hash on the first 40 characters of a sequence comment
		var seq Seq
		if s.err != nil {
			err = fmt.Errorf("reading from file %s: %v, seq num: %d", fname, s.err, nc)
			return
		}
		if seq, err = lump_split(s.b, white[:], s); err != nil {
			err = fmt.Errorf("splitting sequence error: %v\nWorking on seq num: %d, %s", err, nc, s.b)
			return
		}
		var mini int
		if len(seq.cmmt) < h_len {
			mini = len(seq.cmmt)
		} else {
			mini = h_len
		}
		t := seq.cmmt[:mini] // Hash on first h_len characters to look for duplicates
		if v, exists := seq_map[t]; exists {
			s_old := seq_out[v]
			if bytes.Equal(s_old.get_seq(), seq.get_seq()) {
				n_dup++
				if s_opts.Vbsty > 5 {
					fmt.Printf(dup_warn, t, fname)
				}
				continue
			} else {
				if s_opts.Vbsty > 3 {
					fmt.Printf("Likely overlap with %s from %s\n", t, fname)
				}
			}
		}
		seq_map[t] = len(seq_out) // Store the index of this sequence for future comparisons
		seq_out = append(seq_out, seq)
	}
	s.cleanup()
	return
}

// check_seq_lengths() should only be called if we are keeping
// gaps. Then we imagine all the sequences are aligned, so they
// must be the same length. Just print out up to five warnings

func check_lengths(seq_set []Seq) {
	n_warn := 0
	for i, s := range seq_set {
		isiz := s.size()
		for j := i + 1; j < len(seq_set); j++ {
			if jsiz := seq_set[j].size(); isiz != jsiz {
				n_warn++
				fmt.Fprintln(os.Stderr, "We are keeping gaps, but two sequences have different sizes")
				fmt.Fprintf(os.Stderr, "%s has %d sites and %s has %d\n", s.Gene_id(), s.size(),
					seq_set[j].Gene_id(), seq_set[j].size())
				if n_warn >= 5 {
					fmt.Fprintln(os.Stderr, "Not checking seq length any more")
					return
				}
			}
		}
	}
}

// Readfiles takes a slice of filenames and reads sequences from
// each in turn. It returns a slice of type Seq and an error.
func Readfiles(fname []string, s_opts *Options) (seq_set []Seq, n_dup int, err error) {
	seq_map := make(map[string]int)
	seq_set = make([]Seq, 0, 0)
	for _, f := range fname {
		n_dup_onefile := 0
		if seq_set, n_dup_onefile, err = readSeqs(f, seq_set[:], seq_map, s_opts); err != nil {
			return seq_set, n_dup, fmt.Errorf("file %s: %v", f, err)
		}
		n_dup += n_dup_onefile
	}
	if s_opts.Keep_gaps {
		check_lengths(seq_set)
	}
	return
}

// Write_to_f takes a filename and a slice of sequences.
// It writes the sequences to the file.
// For each sequence, it should check if the sequence has been
// set to nil.
// What I could change: If we are removing gaps, we make a buffer which grows
// character by character via WriteByte(). I could make a buffer beforehand
// and grow as necessary.
func Write_to_f(outseq_fname string, seq_set []Seq, s_opts *Options) (err error) {
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
		fmt.Fprintf(outfile_fp, "%c%s\n", cmmt_char, seq.get_cmmt())

		s := seq.get_seq()
		if s_opts.Rmv_gaps { // we have to remove gap characters on output
			n := 0
			for i := range s { //    So we start by looking how many non-gap
				if s[i] != gapchar { //  characters there are.
					n++
				}
			}
			if cap(t) < n { // See if our scratch space is big enough
				t = make([]byte, n)
			}

			m := 0
			for i := range s {
				if s[i] != gapchar {
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


// Copy
func (s *Seq) Copy() (t Seq) {
	t = *new (Seq)
	t.cmmt = s.cmmt
	t.Set_seq (s.get_seq())
	return t
}

// String
func (s Seq) String() (t string) {
	if (len(s.cmmt) > 0) {
		t = fmt.Sprintf ("%c%s\n", cmmt_char, s.get_cmmt())
	}
	t += string(s.get_seq())
	return
}

// CharUsed says how many characters a sequence has used. The idea is that
// one can make a guess if we have a dna or protein sequence
func (s *Seq) CharUsed() (n int) {
	var used [256]bool
	for _, c := range s.get_seq() {
		used[c] = true
	}
	
	for _, t := range used {
		if t == true {
			n++
		}
	}
	return n
}
