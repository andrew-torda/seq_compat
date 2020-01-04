// Package mmcif is the main package. This file is for parsing
// atomsite lines and storing coordinates, insertion codes.
package mmcif

import (
	"errors"
	"fmt"
	. "github.com/andrew-torda/goutil/pdb/cmmn"
	"strconv"
	"sync"
)

const bust = -99.0 // Returned for coordinates when something breaks

type cifCol struct {
	cifName string // name in mmcif file, like label_asym_id
	altName string // an alternative, label_asym_id is the alt for auth_asym_id
	n       int8   // most likely place to find the column
}

type acn struct {
	group_PDB,
	id,
	typeSymbol,
	labelAtomId,
	labelAltId,
	labelCompId,
	labelAsymId,
	labelEntityId,
	labelSeqId,
	pdbxPDBInsCode,
	cartnX,
	cartnY,
	cartnZ,
	occupancy,
	BIsoOrEquiv,
	pdbxFormalCharge,
	authSeqId,
	authCompId,
	authAsymId,
	authAtomId,
	pdbxPDBModelNum cifCol
}

const (
	iniCoordLen = 5 // Initially make room for this many residues
)

// checkName looks to see if a certain header, specified by ndx
// contains the label we are looking for. It makes the header (came
// from a file) lowercase, so we are case insensitive.
func checkName(headers []bSlice, cf cifCol) bool {
	h := headers[cf.n]
	h = sliceAfterASite(h)
	if string(h) == cf.cifName {
		return true
	}
	return false
}

// isDotOrQ returns true if the string is a dot or question mark
func isDotOrQ(s bSlice) bool {
	const (
		dot   = '.'
		qmark = '?'
	)
	if len(s) == 0 {
		return false
	}
	if len(s) == 1 {
		if s[0] == dot || s[0] == qmark {
			return true
		}
	}
	return false
}

// dflt_headers checks if the column names are the likely, default ones from the
// protein data bank. Usually they are, so we do not have to search for
// each label in the table.
func dflt_headers(acn *acn, headers []bSlice) bool {
	if checkName(headers, acn.authAtomId) &&
		checkName(headers, acn.pdbxPDBModelNum) &&
		checkName(headers, acn.cartnX) &&
		checkName(headers, acn.cartnY) &&
		checkName(headers, acn.cartnZ) {
		return true
	}
	return false
}

// boringAtom returns true if the atom name is not on our list of
// interesting atoms. In the bmark.. file, I tested a simple string
// comparison, map / hash and this method. It is fastest, but only
// for short strings and a small set of interesting atoms.
func boringAtom(atName atName, intrstAtoms []atName) bool {
	for _, s := range intrstAtoms { // loop over the candidate names
		if len(atName) != len(s) {
			continue
		}
		switch len(s) {
		case 1:
			if atName[0] == s[0] {
				return false
			}
		case 2:
			if atName[0] == s[0] && atName[1] == s[1] {
				return false
			}
		case 3:
			if atName[0] == s[0] && atName[1] == s[1] && atName[2] == s[2] {
				return false
			}
		default:
			if atName == s {
				return false
			}
		}
	}
	return true
}

// getint16 is a helper used to convert a string to int16 with error messages.
func getint16(toparse bSlice, name string) (int16, error) {
	if len(toparse) == 0 {
		return bust, errors.New("zero length string")
	}
	r, err := strconv.ParseInt(string(toparse), 10, 16)
	if err != nil {
		return bust, errors.New(err.Error() + ". Looked for " + name)
	}
	return int16(r), nil
}

type fswtch byte

const (
	swtchModel fswtch = 1 << iota
	swtchChainID
	swtchAtomID
)

// getxyz gets the x, y and z coordinates from an input line.
// It checks if successful, but this is a bit tedious, so we
// bundle it up into a local function which flags the first
// error. Subsequent calls are no-ops if an error has already occurred.
func getxyz(cmpnt []bSlice, acn *acn) (Xyz, error) {
	var err error
	ff := func(index int8) float32 {
		var xx float64
		if err != nil { // If an error has already occurred,
			return bust //  go no further.
		} // Next line uses a surprising amount of memory.
		if xx, err = strconv.ParseFloat(string(cmpnt[index]), 32); err != nil {
			return bust
		}
		return float32(xx)
	}
	var xyz Xyz
	xyz.X = ff(acn.cartnX.n) // If an error occurred, it will be saved and
	xyz.Y = ff(acn.cartnY.n) // passed back
	xyz.Z = ff(acn.cartnZ.n)
	return xyz, err
}

// sliceAfterAsite takes a string and returns a slice that starts
// after the string, "atom_site.". It gets its own function since it
// is called in two places
func sliceAfterASite(s bSlice) bSlice {
	const slen = len("atom_site.") + 1
	return s[slen:]
}

// getColPos is called if the default cif column names were not correct
// If a name is not found, we do not return an error. We set the
// error that was given to us.
func (cf *cifCol) getColPos(headers []bSlice, err *error) {
	if *err != nil {
		return
	}
	const slen = len("atom_site.") + 1
	for i, h := range headers {
		hh := sliceAfterASite(h)
		if cf.cifName == string(hh) {
			cf.n = int8(i)
			return
		}
	}

	*err = errors.New("Could not find atomsite column: " + cf.cifName)
	return
}

// searchColNames is used when we seem to have something other than defaults.
// It looks in the list of headers for each of the labels we are interested in.
// If ef.findHeader trips an error, subsequent calls will be no-ops, so we just
// check once at the end if an error occurred.
func searchColNames(acn *acn, headers []bSlice) error {
	var err error
	acn.authAtomId.getColPos(headers, &err)
	acn.authCompId.getColPos(headers, &err)
	acn.authAsymId.getColPos(headers, &err)
	acn.authSeqId.getColPos(headers, &err)
	acn.pdbxPDBInsCode.getColPos(headers, &err)
	acn.cartnX.getColPos(headers, &err)
	acn.cartnY.getColPos(headers, &err)
	acn.cartnZ.getColPos(headers, &err)
	acn.pdbxPDBModelNum.getColPos(headers, &err)
	return err
}

// See if two chain names are the same. 99% of the time, these are
// one letter strings, so we can do a special comparison
func chainEq(a chainID, b chainID) bool {
	if len(a) == 1 && len(b) == 1 { // 99 % of the time we are here
		if a[0] == b[0] {
			return true
		}
		return false
	}
	if a.string() == b.string() {
		return true
	}
	return false
}

// chanWrap wraps a channel and stores the slice we get from it.
type chanWrap struct { // It lets us go line by line and
	c       chan []bSlice //   allows pushback.
	cs      []bSlice      // The slice of bit slices with our strings
	bufPool *sync.Pool    // Pool created in the caller and shared here
	scrtch  [40]bSlice    // Scratch space
	ndx     int           //  Using index, instead of reslicing lets us do pushback
}

// linechan returns the next line from the channel which has slices of lines.
func (cw *chanWrap) linechan() bSlice {
	if cw.ndx == len(cw.cs) { // refill
		if len(cw.cs) > 0 {
			cw.bufPool.Put(cw.cs)
		}
		cw.cs = <-cw.c
		cw.ndx = 0
		if len(cw.cs) == 0 {
			return nil
		}
	}
	cw.ndx++
	return cw.cs[cw.ndx-1]
}

// closePool puts buffered space back in the pool
func (cw *chanWrap) closePool() {
	cw.bufPool.Put(cw.cs)
}

// cmpntChan calls linechan to get the next line of input and
// returns it, broken into components.
func (cw *chanWrap) cmpntChan() (cmpnt []bSlice) {
	s := cw.linechan()
	if s == nil {
		return nil
	}
	cmpnt = fields(s, cw.scrtch[:])
	for i := range cmpnt {
		s = cmpnt[i]  // A string like "c5'" comes to us with unwanted quotes
		first := s[0] // around it
		last := s[len(s)-1]
		if  first == dquote && last == dquote ||
			first == squote && last == squote {
			if len(s) > 1 {
				cmpnt[i] = s[1:len(s) - 1]
			}
		}
	}
	return cmpnt
}

// pushback pushes a line from a channel back in the wrapper structure.
func (cw *chanWrap) pushback() {
	if cw.ndx == 0 {
		return
	}
	cw.ndx--
}

// rsdue is the first part of parsing an atom and passed back by getRsdue
type rsdue struct {
	resType bSlice
	atName  atName
	numLbl  int // Residue number as label. Can be anything, auth_seq_id
	chainID chainID
	mdlNum  int16
	insCode byte
}

// cmpntTooSmall checks if our compontent array is too small
func cmpntTooSmall(acn *acn, len int) error {
	msg := "Too few components (%d)"
	i := int8(len)
	if i <= acn.pdbxPDBModelNum.n {
		return fmt.Errorf(msg, len)
	}
	if i <= acn.authAtomId.n {
		return fmt.Errorf(msg, len)
	}
	if i <= acn.cartnZ.n {
		return fmt.Errorf(msg, len)
	}
	return nil
}

// getAtom returns the string for the atom name
func getatom(cmpnt []bSlice, acn *acn) atName {
	return atName(cmpnt[acn.authAtomId.n])
}

// getResnumStr returns the string for the residue number. The following
// function will return the number form.
func getResnumStr(cmpnt []bSlice, acn *acn) bSlice {
	return cmpnt[acn.authSeqId.n]
}

// getResnumNum returns the residue number, but as an int.
func getResnumNum(cmpnt []bSlice, acn *acn) (int, error) {
	s := getResnumStr(cmpnt, acn)
	if isDotOrQ(s) {
		return BrokenResNum, nil
	}

	if t, err := strconv.ParseInt(string(s), 10, 32); err != nil {
		err = fmt.Errorf("%s: Converting residue number %s", err.Error(), s)
		return -1, err
	} else {
		return int(t), nil
	}
}

func getInsCode(cmpnt []bSlice, acn *acn) (ret byte, err error) {
	t := cmpnt[acn.pdbxPDBInsCode.n]
	if isDotOrQ(t) {
		return
	}
	if len(t) > 1 {
		return ret, errors.New("insertion code length > 1: \"" + string(t) + "\"")
	}
	return t[0], nil
}

func getChainID(cmpnt []bSlice, acn *acn) chainID {
	s := cmpnt[acn.authAsymId.n]
	if !isDotOrQ(s) {
		return (chainID(s))
	}
	return ""
}

// func getMdlNumstr returns the model number as a string. This
// is enough if we are just looking to see if the number changed.
// The succeeding function does the conversion to an int.
func getMdlNumstr(cmpnt []bSlice, acn *acn) bSlice {
	return cmpnt[acn.pdbxPDBModelNum.n]
}

// func getMdlNumNum converts to an integer.
// There is something arbitrary here. If we have a dot or a question
// mark, we do not return error. We return a broken residue number.
func getMdlNumNum(cmpnt []bSlice, acn *acn) (mdlNum int16, err error) {
	s := getMdlNumstr(cmpnt, acn)
	if mdlNum, err = getint16(s, "model num"); err != nil {
		msg := "\nFull string was"
		for _, x := range cmpnt {
			msg = msg + " " + string(x)
		}
		err = errors.New(err.Error() + msg)
		return -1, err
	}
	return mdlNum, nil
}

// getRsdue takes a line and returns information about the residue
func getRsdue(cmpnt []bSlice, acn *acn, rsdue *rsdue) error {
	var err error
	if rsdue.numLbl, err = getResnumNum(cmpnt, acn); err != nil {
		return err
	}
	if rsdue.insCode, err = getInsCode(cmpnt, acn); err != nil {
		return err
	}

	rsdue.resType = cmpnt[acn.authCompId.n]
	rsdue.chainID = getChainID(cmpnt, acn)
	rsdue.atName = getatom(cmpnt, acn)
	if rsdue.mdlNum, err = getMdlNumNum(cmpnt, acn); err != nil {
		return err
	}

	return nil
}

// A onemodel is basically just coordinates. When we read from a file,
// we get insertion codes, residue names, ...
// Here is a structure which lets us return the whole lot
type modelextra struct {
	mdltmp  *onemodel
	insCode []byte
	resname []bSlice
	numLbl  []int
	chainID chainID
	err     error
	mdlNum  int16
	status  byte
}

const ( // Values used by nxtMdl to report what it found
	mEmpty byte = iota // No problem, but no match on model/chain
	mOk                // Got data. Happy
	mEOF               // No error, but end of input data.
	mErr               // Error occurred
)

// Set up a byte which will tell us if we have to switch according
// to model, chain or atomtype
func fswtchSet(fltr *fltr) (fswtch fswtch) {
	if fltr.modelMax >= 0 {
		fswtch |= swtchModel
	}
	if len(fltr.chains) > 0 {
		fswtch |= swtchChainID
	}
	if len(fltr.intrstAtoms) > 0 { // XXXX and fltr.intrstAtoms[0] != ""
		fswtch |= swtchAtomID
	}
	return fswtch
}

// boringLine determines if an input line is boring and of no interest.
// It might be called to skip over a zillion unwanted lines, so it should
// quickly determine if it can return true - usually if the model or
// chain is not wanted.
func boringLine(cmpnt []bSlice, acn *acn, fltr *fltr, fswtch fswtch) bool {
	if fswtch&swtchModel != 0 {
		nMdl, err := getMdlNumNum(cmpnt, acn)
		if err != nil {
			return true
		}
		if nMdl > fltr.modelMax {
			return true
		}
	}
	if fswtch&swtchChainID != 0 {
		for _, wanted := range fltr.chains {
			if chainEq(wanted, getChainID(cmpnt, acn)) {
				goto chaingood
			}
		}
		return true // Boring because chain is not wanted
	}
chaingood:
	if fswtch&swtchAtomID != 0 { // jump over unwanted atoms at start
		atName := getatom(cmpnt, acn)
		if boringAtom(atName, fltr.intrstAtoms) {
			return true
		}
	}
	return false
}

// getMdl reads one model. If it is interesting, it returns it.
// It can, of course, just return empty if there is nothing to
// find.
// There is a line, 			mEx.resname = append(mEx.resname, rsdue.resType)
// which costs much memory. I think there is a solution. Store an index into a table of
// residue type names. The table can be initialised with amino acids and nucleotides
// we know about. If we encounter a new one, we add it. We need a corresponding map
// to go from names to numbers. This can be initialised by walking down the array.
func nxtMdl(cw *chanWrap, headers []bSlice, fltr *fltr,
	fswtch fswtch, acn *acn) *modelextra {
	mExEOF := &modelextra{status: mEOF}
	var cmpnt []bSlice
	for cmpnt = cw.cmpntChan(); ; cmpnt = cw.cmpntChan() {
		if len(cmpnt) == 0 { // an unteresting model or chain
			return mExEOF // or from unwanted atoms
		}
		if cmpnt == nil {
			return &modelextra{status: mErr, err: errors.New("no components")}
		}
		if err := cmpntTooSmall(acn, len(cmpnt)); err != nil {
			err = fmt.Errorf("%s on line %v", err.Error(), cmpnt)
			return &modelextra{status: mErr, err: err}
		}
		if !boringLine(cmpnt, acn, fltr, fswtch) {
			break
		}
	}

	var haveAtomList bool
	if (fswtch & swtchAtomID) != 0 {
		haveAtomList = true
	} else {
		haveAtomList = false
	}

	// Now we have a model we want and a chain we want. Save model
	mdltmp := make(onemodel)
	if (fswtch & swtchAtomID) != 0 {
		for _, at := range fltr.intrstAtoms {
			mdltmp[at] = make(XyzSl, 0, iniCoordLen)
		}
	}
	mEx := &modelextra{status: mOk}
	mEx.insCode = make([]byte, 0, iniCoordLen)
	mEx.resname = make([]bSlice, 0, iniCoordLen)
	mEx.numLbl = make([]int, 0, iniCoordLen)
	mEx.chainID = getChainID(cmpnt, acn)

	mEx.mdlNum, _ = getMdlNumNum(cmpnt, acn)
	var oldRsdue rsdue
	if err := getRsdue(cmpnt, acn, &oldRsdue); err != nil {
		err = errors.New(err.Error() + " Broke, 1st residue")
		return &modelextra{status: mErr, err: err}
	}
	for ; cmpnt != nil; cmpnt = cw.cmpntChan() {
		if err := cmpntTooSmall(acn, len(cmpnt)); err != nil {
			err = fmt.Errorf("%s on line %v", err.Error(), cmpnt)
			return &modelextra{status: mErr, err: err}
		}

		newres := false
		var rsdue rsdue
		if err := getRsdue(cmpnt, acn, &rsdue); err != nil {
			return &modelextra{status: mErr, err: err}
		}
		if haveAtomList && boringAtom(rsdue.atName, fltr.intrstAtoms) {
			continue
		}
		// Maybe finished with this model
		if rsdue.mdlNum != mEx.mdlNum || rsdue.chainID != mEx.chainID {
			cw.pushback()
			break
		}

		// Next line asks if we need to expand the per-residue arrays
		// Perversity:
		if rsdue.insCode != oldRsdue.insCode ||
			rsdue.numLbl != oldRsdue.numLbl ||
			(cmpnt[0][0] == 'H' && (oldRsdue.atName == rsdue.atName)) {
			newres = true
		}
		if len(mEx.insCode) == 0 {
			newres = true
		}

		if !haveAtomList {
			if _, ok := mdltmp[rsdue.atName]; !ok {
				li := len(mEx.insCode)
				mdltmp[rsdue.atName] = make(XyzSl, li, li)
				for i := 0; i < li; i++ {
					mdltmp[rsdue.atName][i] = BrokenXyz
				}
			}
		}
		if newres {
			for at, _ := range mdltmp {
				mdltmp[at] = append(mdltmp[at], BrokenXyz)
			}
			mEx.insCode = append(mEx.insCode, rsdue.insCode)
			mEx.resname = append(mEx.resname, rsdue.resType) // very memory expensive
			mEx.numLbl = append(mEx.numLbl, rsdue.numLbl)    // relatively expensive
			oldRsdue = rsdue
		}
		if xyz, err := getxyz(cmpnt, acn); err != nil {
			return &modelextra{status: mErr, err: err}
		} else {
			i := len(mEx.insCode) - 1
			mdltmp[rsdue.atName][i] = xyz
		}
	}
	mEx.mdltmp = &mdltmp
	mEx.status = mOk
	return mEx
}

// addmodel adds a model we have read to the existing data
func (md *MmcifData) addmodel(mEx *modelextra) error {
	var thisChain oneChain
	if value, ok := md.Allcoord[mEx.chainID]; ok != true {
		thisChain = newOneChain()
	} else {
		thisChain = value
	}
	thisChain.coords = append(thisChain.coords, *mEx.mdltmp)
	if len(thisChain.insCode) == 0 {
		thisChain.insCode = mEx.insCode
		thisChain.numLbl = mEx.numLbl
	} else { // there used to be a test for consistent model lengths.
		// but it there is a tiny number of files where the number
	} // of atoms changes from model to model

	md.Allcoord[mEx.chainID] = thisChain
	return nil
}

// fillme does the work reading atom_site lines
func (md *MmcifData) fillme(c chan []bSlice, headers []bSlice, fltr *fltr,
	fswtch fswtch, acn *acn, bufPool *sync.Pool) error {
	cw := &chanWrap{c: c, bufPool: bufPool}
	defer cw.closePool()
	mEx := nxtMdl(cw, headers, fltr, fswtch, acn)
	for ; mEx.status != mEOF; mEx = nxtMdl(cw, headers, fltr, fswtch, acn) {
		if mEx.status == mErr {
			return mEx.err
		}
		if mEx.status == mEmpty {
			continue
		} // We found some information, so put it in the structure to return
		if err := md.addmodel(mEx); err != nil {
			return err
		}
	}
	return nil
}

// drain discards anything in the channel
func drain(c chan []bSlice) {
	for _ = range c {
	}
}

// atomSite reads lines of input from the channel, but it gets a
// few of them at once - a slice is fed into the channel.
// When it starts, the first line is the set of names.
// For each interesting atom type, we set up a xyz slice in the
// coord hash.
func atomSite(headers []bSlice, fltr *fltr, md *MmcifData,
	c chan []bSlice, rChan chan string, bufPool *sync.Pool) {
	acn := &acn{
		cifCol{"group_PDB", "", 0},
		cifCol{"id", "", 1},
		cifCol{"type_symbol", "", 2},
		cifCol{"label_atom_id", "", 3},
		cifCol{"label_alt_id", "", 4},
		cifCol{"label_comp_id", "", 5},
		cifCol{"label_asym_id", "", 6},
		cifCol{"label_entity_id", "", 7},
		cifCol{"label_seq_id", "", 8},
		cifCol{"pdbx_PDB_ins_code", "", 9},
		cifCol{"Cartn_x", "", 10},
		cifCol{"Cartn_y", "", 11},
		cifCol{"Cartn_z", "", 12},
		cifCol{"occupancy", "", 13},
		cifCol{"B_iso_or_equiv", "", 14},
		cifCol{"pdbx_formal_charge", "", 15},
		cifCol{"auth_seq_id", "label_seq_id", 16},
		cifCol{"auth_comp_id", "label_comp_id", 17},
		cifCol{"auth_asym_id", "label_asym_id", 18},
		cifCol{"auth_atom_id", "label_atom_id", 19},
		cifCol{"pdbx_PDB_model_num", "", 20},
	}

	defer close(rChan)

	md.Allcoord = make(map[chainID]oneChain)

	fswtch := fswtchSet(fltr)
	if !dflt_headers(acn, headers) { // Check if we have default pdb headers
		if err := searchColNames(acn, headers); err != nil {
			drain(c)
			rChan <- err.Error()
			return
		}
	}
	if err := md.fillme(c, headers, fltr, fswtch, acn, bufPool); err != nil {
		drain(c)
		rChan <- err.Error()
		return
	}
	return
}
