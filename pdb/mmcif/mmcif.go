// Package mmcif reads an mmcif formatted file. It is a subpackage of pdb.
// The first thing to do is build an mmcifreader and then call it.
package mmcif

import (
	"bufio"
	"bytes"
	"github.com/andrew-torda/goutil/pdb/cmmn"
	"errors"
	"fmt"
	"io"
	"sync"
)

const (
	squote byte = '\''
	dquote byte = '"'
)

// Usually one reads a file which contains lots of information you are
// not interested in. Of the stuff you keep, like coordinates, only some
// of them are interesting. We handle this in two states.
// 1. Make a list of interesting data items and tables. If something is
// not on this list, do not save it.
// 2. If something is potentially interesting, like the table of coordinates
// put it in this structure. You can then pick from here.

type bSlice []byte // byte slice
type stSlice []string
type keepTable struct {
	Names []string  // table headings
	Vals  []stSlice // each entry is a slice of values
}

type chainID string

func (c chainID) string() string { return string(c) }

type atName string

func (a atName) string() string { return string(a) }

type onemodel map[atName]cmmn.XyzSl
type oneChain struct {
	coords  []onemodel
	numLbl  []int
	insCode []byte
}

// a set of coordinates for one chain goes into
type stringhash map[string]string
type tablehash map[string]keepTable

// MmcifData defines the data that will be returned by an mmcifReader
type MmcifData struct { // This is a stuttering name. Should be data.
	Data     stringhash // Data items to keep
	Tables   tablehash  // Tables we keep
	Allcoord map[chainID]oneChain
}

func newOneChain() oneChain {
	return oneChain{
		coords: make([]onemodel, 0, 1),
	}
}

// What criteria do we use when deciding whether or not to keep
// an atom
type fltr struct {
	modelMax    int16 // How many models should be read ?
	chains      []chainID
	intrstAtoms []atName // Interesting atoms - those we keep
}

// MmcifReader is the object which will do the reading of mmcif data
// We do not return information here. Here is where we store instructions
// to the reader.
type MmcifReader struct { // This is also a stuttering name
	cmmtScanner
	dataToKeep   map[string]bool
	tablesToKeep map[string]bool
	fltr         *fltr
	headers      []bSlice
	scrtchBytes  [][]byte
	nmodel       uint16 // how many models do I want to keep
}

// Dump is for debugging. Dump out what we have stored
func (md *MmcifData) Dump() {
	for k, v := range md.Data {
		fmt.Println(k, ":", v)
	}
	for k, v := range md.Tables {
		fmt.Println("m.tables", k, ":", v)
	}
}

// NewMmcifReader returns an object to read mmcif files.
// It is given a reader, so the caller must have decided if it is
// a file, compressed file, http source, whatever.
// We do have a list of hard-coded default atoms. I don't know if this
// is a good or bad idea.
// If one wants to default to chain "A", say chains := []chainID{"A"}
func NewMmcifReader(r io.Reader) *MmcifReader {
	if r == nil {
		return nil
	}
	intrstAtoms := []atName{"CA", "C", "CB", "N", "O"} // just a default set

	var chains []chainID
	return &MmcifReader{
		cmmtScanner:  newCmmtScanner(r, '#'),
		dataToKeep:   make(map[string]bool),
		tablesToKeep: make(map[string]bool),
		fltr:         &fltr{modelMax: 5, chains: chains, intrstAtoms: intrstAtoms},
		scrtchBytes:  make([][]byte, 25),
		nmodel:       1,
	}
}

// SetChains adds a list of desired chains to our reader.
func (mr *MmcifReader) SetChains(s []string) {
	mr.fltr.chains = make([]chainID, 0)
	for _, c := range s {
		if c != "" {
			mr.fltr.chains = append(mr.fltr.chains, chainID(c))
		}
	}
}

// SetAtoms sets the slice of atom names which we will look for
func (mr *MmcifReader) SetAtoms(s []string) {
	mr.fltr.intrstAtoms = make([]atName, len(s), len(s))
	for i := range s {
		mr.fltr.intrstAtoms[i] = atName(s[i])
	}
}

// AddItems adds a category to the reader that we will read from
// the mmcif file.
func (mr *MmcifReader) AddItems(s []string) {
	for _, a := range s {
		mr.dataToKeep[a] = true
	}
}

// AddTable tells us that if we see a table / loop with a certain word as
// the first entry, than we will keep this table.
func (mr *MmcifReader) AddTable(s []string) {
	for _, a := range s {
		mr.tablesToKeep[a] = true
	}
}

// CmmtScanner is a wrapper around bufio.Scanner that will ignore anything
// after a comment character and remove leading and trailing white space.
// It also counts newlines in scanner.n, so we can print out the line
// number in error messages.
type cmmtScanner struct {
	*bufio.Scanner           // standard library scanner
	l_err          readError // fill this out as soon as an error happens
	ctoken         []byte    // Store the bytes that will be returned by cbytes()
	n              int       // line number in the mmcif file
	cmmt           byte      // Comment character
	Ok             bool      // Are we OK or have we had an error ?
}

// AddModelMax tells us the maximum number of models to read.
// -1 means get everything
//  0 means get nothing
//  a positive in is the number of models
func (mr *MmcifReader) SetModelMax(modelMax int16) {
	mr.fltr.modelMax = modelMax
}

// NewCmmtScanner is a wrapper around scanner, but
//  - jumps over blank lines
//  - removes leading and trailing space
//  - removes anything after a comment character
// An mmcifReader contains a newCmmtScanner.
func newCmmtScanner(r io.Reader, cmmt byte) cmmtScanner {
	return cmmtScanner{
		Scanner: bufio.NewScanner(r), // the other fields get
		cmmt:    cmmt,                // default values (zeroes and empty)
		Ok:      true,
	}
}

// cscan is a wrapper around the library Scan(). It adds a newline counter
// for error messages. It jumps over blank lines and lines starting
// with a comment character. Comment characters are only recognised as the
// first character, since they are legitimate elsewhere in the text.
// When finished, it sets "ctoken" to point to the slice.
func (s *cmmtScanner) cscan() (ok bool) {
	var b []byte
	if !s.Ok { // We have already had an error, but nobody has noticed.
		s.ctoken = nil
		s.fill("pre-existing error missed. Small bug ?", false)
		return false // Just get out of here
	}
	ok = true
	for len(b) == 0 && ok {
		if ok = s.Scan(); ok { //     This is false on EOF
			s.n++ //                  Counter for error messages
		} else { //                   If scan returned false,
			s.ctoken = nil //         but Err() is nil, it is just EOF
			if s.Err() != nil {
				s.fill(s.Err().Error(), true) // This is a real error
				return false
			}
			return true // No error, just EOF
		}
		b = s.Bytes()    //  Get the raw line from library scanner
		if len(b) == 0 { // blank line
			continue
		}
		if b[0] == s.cmmt { // Comment,
			b = nil
		}
	}
	s.ctoken = b
	return ok
}

// cbytes is like Bytes from the library, but returns the processed characters
// (with stuff after the comment removed).
func (s *cmmtScanner) cbytes() []byte {
	return s.ctoken
}

// stateFn is the type of state function. It returns the next
// state function that should act on its input.
type stateFn func(*MmcifReader, *MmcifData) stateFn

// stateData reads lines that start with __DATA
func stateData(mr *MmcifReader, _ *MmcifData) stateFn {
	if !mr.cscan() {
		return nil
	}
	return stateTop
}

// stateUnknown should be reached if we are confused and do not know
// what to do. It is an error and we should stop
func stateUnknown(mr *MmcifReader, _ *MmcifData) stateFn {
	mr.fill("In Unknown state", true)
	return nil
}

// stateLoophdr gets the headers from a loop directive
// It also gets to make a decision about what to do next.
// If the headers are deemed interesting, it calls stateLoopTable
// If headers are for a table that we want to skip, it
// should call stateSkipLoopTable.
func stateLoopHdr(mr *MmcifReader, _ *MmcifData) stateFn {
	if len(mr.headers) != 0 {
		mr.fill("probable bug, headers slice not empty", false)
		return nil
	}
	for ok := true; ok && mr.cbytes()[0] == byte('_'); ok = mr.cscan() {
		s := make([]byte, len(mr.cbytes()))
		copy(s, mr.cbytes())
		t := bytes.TrimRight(s, " ")
		mr.headers = append(mr.headers, t)
	}
	if len(mr.headers) < 1 {
		mr.fill("no contents found while reading loop headers", true)
		mr.headers = mr.headers[:0]
		return nil
	}

	dots := []byte{'.'}
	kword := bytes.SplitAfter(mr.headers[0], dots)
	if bytes.HasPrefix(kword[0], []byte("_atom_site.")) {
		return stateAtomTable
	}
	if _, ok := mr.tablesToKeep[string(kword[0])]; ok { // xxxx
		return stateLoopTable
	}
	mr.headers = mr.headers[:0]
	return stateSkipLoopTable
}

// ispecial returns true if the input in inline is not simply
// more of a table. Usually this means there is a new directive
// coming.
// If we have end of line, we also return true, so a caller knows
// it has to do something special.
// We used to stop if we saw "data", but this is sometimes present in tables
func isSpecial(inline []byte) bool {
	switch {
	case inline == nil:
		return true
	case bytes.HasPrefix(inline, []byte("_")):
		return true
	case bytes.HasPrefix(inline, []byte("loop_")):
		return true
	default:
		return false
	}
}

// stateLoopTable reads from each line a line in a table.
// Build the table within this function and then put it in the hash
// table of tables.
// Collect the output from the function where the println is.
func stateLoopTable(mr *MmcifReader, md *MmcifData) stateFn {
	const not_split string = "Could not split string at dot: "
	dots := []byte{'.'}
	ncol := len(mr.headers)
	var table keepTable
	var tblName string
	{ //             This first section just saves the headers for table
		t := bytes.SplitAfterN(mr.headers[0], dots, 2)
		if len(t) < 2 {
			mr.fill(not_split+string(mr.headers[0]), true)
			return nil
		}
		s := string(t[0])
		tblName = s[0 : len(s)-1]
	}
	table.Names = make([]string, 0, len(mr.headers))
	table.Vals = make([]stSlice, 0, 5) // I just made this up, "5"

	for _, word := range mr.headers { // given _atom_site.foo, save foo
		t := bytes.SplitAfterN(word, dots, 2)
		if len(t) < 2 {
			mr.fill(not_split+string(word), true)
			return nil
		}
		table.Names = append(table.Names, string(t[1]))
	}
	mr.headers = mr.headers[:0] // do not need this copy of the headers any more
	// Finished with headers, now read table contents
	for b, ok := getNpieces(mr, ncol); len(b) == ncol && ok; {
		table.Vals = append(table.Vals, b)
		b, ok = getNpieces(mr, ncol)
	}
	md.Tables[tblName] = table
	return stateTop
}

const line_siz = 92 // A line from PDB is 88 bytes long
const sl_siz = 50   // This comes from bencharking. Set to 50
// newLineBuf creates the slice of lines (byte slices) that are
// filled and used to send information to the reader (atomsite()).
func newLineBuf() interface{} {
	var tmp [sl_siz * line_siz]byte
	var x [sl_siz]bSlice //  Since this is known at compile time
	for i, start, end := 0, 0, line_siz; i < sl_siz; i++ {
		x[i] = tmp[start:end:end]
		start = end
		end += line_siz
	}
	return x[:]
}

// stateAtomTable is like any stateLoopTable, but we special case it
// because it is the biggest, most important and slowest to
// process. There are lots of strings that should be discarded or
// converted.
// We read lines into a slice of lines. When we have enough, we
// push the slice into the channel. AtomSite() does the processing.
// In the meantime, we continue reading the file.
// At the start, we have the same information as stateLoopTable
func stateAtomTable(mr *MmcifReader, md *MmcifData) stateFn {
	var nothing string
	c := make(chan []bSlice, 3) // buffer size 3 came from benchmarking
	rChan := make(chan string)
	// We make new buffers (line slices) here. The other end of the channel
	// puts the buffers back in the pool when it has processed all the lines.
	var bufPool = sync.Pool{
		New: newLineBuf,
	}

	i := 0
	{
		var headers []bSlice
		for _, h := range mr.headers {
			tmp := make([]byte, len(h))
			copy(tmp, h)
			headers = append(headers, tmp)
		}
		go atomSite(headers, mr.fltr, md, c, rChan, &bufPool)
	}

	mr.headers = nil // do no need them any more
	strings := bufPool.Get().([]bSlice)
	for {
		t := mr.cbytes()
		if len(t) > cap(strings[i]) { // We only reach here if the default line
			strings[i] = make([]byte, len(t)) // length is too small.
		}
		strings[i] = strings[i][:len(t)]
		copy(strings[i], t)

		if i == (sl_siz - 1) {
			i = 0                              // send the accumulated lines into
			c <- strings                       // the channel (to atomSite()) and
			strings = bufPool.Get().([]bSlice) // get fresh storage space from the pool
			for i := range strings {
				strings[i] = strings[i][:0]
			}
		} else {
			i++
		}
		if !mr.cscan() {
			break
		}
		s := mr.cbytes()
		if isSpecial(s) {
			break
		}
	}
	if i > 0 { // Push any leftover strings down the channel
		c <- strings[0:i]
	}
	close(c)
	if s := <-rChan; s != nothing {
		mr.fill(s, false) // On this channel, any non-zero
		return nil        // is an error
	}
	return stateTop
}

// stateSkipLoopTable reads lines from a table, but does not
// save them anywhere. Most of the tables we encounter are not
// to be saved.
func stateSkipLoopTable(mr *MmcifReader, _ *MmcifData) stateFn {
	found_something := false
	for ; !isSpecial(mr.cbytes()); mr.cscan() {
		found_something = true
	}
	if !found_something {
		mr.fill("empty table", true)
		return nil
	}
	return stateTop
}

// stateLoop is where you are if you have a loop directive.
// You just have to jump over the line and go to reading the
// headers.
func stateLoop(mr *MmcifReader, _ *MmcifData) stateFn {
	if !mr.cscan() {
		return nil
	}
	return stateLoopHdr
}

// stateDItem gets a data item. This is often on one line, but
// if rest has zero length, the value is on subsequent lines
func stateDItem(mr *MmcifReader, md *MmcifData) stateFn {
	var value string
	var is_interesting_item = false
	t, err := splitCifLine(mr.cbytes(), mr.scrtchBytes)

	if err != nil {
		mr.fill(err.Error(), true)
		return nil
	}

	itemName := t[0]
	if _, present := mr.dataToKeep[string(itemName)]; present {
		is_interesting_item = true
	}
	if len(t) == 2 && t[1][0] != ';' { // Simplest. We just have a value on the line
		value = string(t[1])
		if !mr.cscan() {
			mr.fill("looking for data item", true)
			return nil
		}
	} else if len(t) == 1 {
		const msg string = "data split on two lines"
		if !mr.cscan() {
			mr.fill(msg, true)
			return nil
		}
		b_in := mr.cbytes()
		if b_in[0] == ';' {
			ok := true
			tmp := string(b_in[1:])
			for ok = mr.cscan(); len(mr.cbytes()) > 0 && ok; ok = mr.cscan() {
				if mr.cbytes()[0] == ';' {
					break
				}
				tmp = tmp + string(mr.cbytes())
			}
			if !ok {
				mr.fill(msg, true)
				return nil
			}
			value = tmp
		} else {
			value = string(b_in)
		}
		mr.cscan() // If an error occurs, the next function will pick it up
	} else {
		panic (fmt.Sprintf ("prog bug, line %d, buf %s\n", mr.n, string(mr.cbytes())))
	}

	if is_interesting_item {
		md.Data[string(itemName)] = value
	}
	return stateTop
}

// stateTop is the general state that looks at the current line and
// decides what state to jump to next.
func stateTop(mr *MmcifReader, _ *MmcifData) stateFn {
	b := mr.cbytes() // Does not advance scanner
	if mr.Ok != true {
		return nil
	}
	sData := []byte("data")
	sDItem := []byte("_") // a "data item"
	sLoop := []byte("loop_")
	switch {
	case b == nil:
		return nil
	case bytes.HasPrefix(b, sLoop):
		return stateLoop
	case bytes.HasPrefix(b, sData):
		return stateData
	case bytes.HasPrefix(b, sDItem):
		return stateDItem
	default:
		return stateUnknown
	}
}

// getNpieces asks the scanner for lines and returns N items
// as an array of strings. We have to use new strings, since
// calls to scan() will update the underlying buffer.
func getNpieces(mr *MmcifReader, npiece int) (ret []string, ok bool) {
	// notNasty returns true if we cannot use the library split
	// function. That is, it contains quotes.
	notNasty := func(b_in []byte) bool {
		for _, c := range b_in {
			if c == dquote || c == squote {
				return false
			}
		}
		return true
	}
	for ok = true; len(ret) < npiece && ok; ok = mr.cscan() {
		b_in := mr.cbytes()
		if isSpecial(b_in) {
			return nil, ok
		}
		if b_in[0] == ';' {
			tmp := string(b_in[1:])
			for ok = mr.cscan(); ok; ok = mr.cscan() {
				x := mr.cbytes()
				if len(x) < 1 || x[0] == ';' {
					break
				}
				tmp = tmp + string(x)
			}
			ret = append(ret, tmp)
			if !ok {
				mr.fill("getNpieces", true)
				return nil, false
			}
			continue
		} else {
			var t [][]byte
			if notNasty(b_in) { //       For a clean string, just
				t = bytes.Fields(b_in) // use library function
			} else {
				var err error
				if t, err = splitCifLine(b_in, mr.scrtchBytes); !ok {
					mr.fill(err.Error(), true)
					return nil, false
				}
			}
			for _, u := range t {
				ret = append(ret, string(u))
			}
		}
	}
	return
}

// DoFile needs a better name.
// This takes an mmcifreader and actually parses the file.
func (mr *MmcifReader) DoFile() (*MmcifData, error) {
	if mr == nil {
		return nil, errors.New("Start of file, nil mmcifReader")
	}
	if !mr.cscan() {
		return nil, mr.l_err
	}
	md := new(MmcifData)        // Coordinates, insertion codes
	md.Data = make(stringhash)  // and data from the atom_sites
	md.Tables = make(tablehash) // is set up when we encounter it.
	md.Allcoord = nil           // Make this clear. It is allocated when we use it.
	for state := stateTop; (state != nil) && mr.Ok; {
		state = state(mr, md)
	}
	if !mr.Ok {
		return nil, mr.l_err
	}
	if mr.n == 0 {
		mr.fill("zero length file", false)
	}

	if mr.Ok != true {
		return nil, mr.l_err
	}

	for k := range mr.dataToKeep {
		if _, ok := md.Data[k]; !ok {
			break
		}
	}
	return md, nil
}

// Some functions for putting into tests.
// function nAtomtype returns the number of atoms in a set
// of coordinates, summed over all chains and all models
func (md *MmcifData) NAtomAllChainAll() (valid, invalid int) {
	if md == nil {
		return
	}
	for _, onechain := range md.Allcoord { // _ is the chainID
		for _, model := range onechain.coords { // _ is model index
			for _, xyzS := range model { // _ is ??
				for _, xyz := range xyzS { // _ is atomname
					if xyz.Ok() {
						valid++
					} else {
						invalid++
					}
				}
			}
		}

	}
	return
}

// nAtomType gives us the number of atoms of a specific type, summed
// over all models and chains
func (md *MmcifData) nAtomType(a string) (n int) {
	for _, onechain := range md.Allcoord {
		for _, model := range onechain.coords {
			xyzS := model[atName(a)]
			for _, x := range xyzS {
				if x.Ok() {
					n++
				}
			}
		}
	}
	return n
}

// getXyz returns a slices of xyz's given a model, chain and atom type
func (md *MmcifData) getXyz(mdlNum int16, chn string, at string) (cmmn.XyzSl, error) {
	var err error
	atName := atName(at)
	chainID := chainID(chn)
	mdlerr := "Wanted model num %d, but last one is indexed with %d"
	if _, ok := md.Allcoord[chainID]; !ok {
		return nil, fmt.Errorf("no chain %s", chn)
	}
	onechain := md.Allcoord[chainID]
	if int(mdlNum) >= len(onechain.coords) {
		return nil, fmt.Errorf(mdlerr, mdlNum, len(onechain.coords)-1)
	}
	mdl := onechain.coords[mdlNum]
	if _, ok := mdl[atName]; !ok {
		return nil, fmt.Errorf("No atom of type %s", atName)
	}
	return md.Allcoord[chainID].coords[mdlNum][atName], err
}

// floatDiff returns true if the difference between two float numbers
// is greater than tolerance, which is hard-coded.
func floatDiff(a, b float32) bool {
	const tol float32 = 0.05
	d := a - b
	if d > tol || -d > tol {
		return true
	}
	return false
}

// XyzNotEq is used in the tests. We need to do floating point
// comparisons. If two points differ in x, y or z by more than "tol",
// we return true (not equal). Tol is arbitrarily set at 0.05 since
// we are interested in gross errors, not numerical detail.
func XyzNotEq(x1 []float32, x2 cmmn.Xyz) bool {
	if floatDiff(x1[0], x2.X) {
		return true
	}
	if floatDiff(x1[1], x2.Y) {
		return true
	}
	if floatDiff(x1[2], x2.Z) {
		return true
	}
	return false
}
