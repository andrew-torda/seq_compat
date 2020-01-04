// This is the upper level for reading PDB files.
// Decide if a file is compressed or not, and what format
// we are going to read. Then call the corresponding pdb or mmcif
// format reader.

package pdb

import (
	"github.com/andrew-torda/goutil/pdb/cmmn"
	"github.com/andrew-torda/goutil/pdb/mmcif"
	"github.com/andrew-torda/goutil/pdb/zwrap"
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	old_fmt byte = iota
	mmcif_fmt
	unk_fmt
)

// a fakecloser is a wrapper around a io.Writer which turns it into
// a WriteCloser.
type fakecloser struct {
	io.Writer
}

func (fakecloser) Close() error { return nil }

// comparefirst looks says if two words are the same, looking at the
// the length of the shorter
func comparefirst(s, t string) bool {
	l1 := len(s)
	l2 := len(t)
	if l2 < l1 {
		l1 = l2
	}
	s = s[:l1]
	t = t[:l2]
	if s == t {
		return true
	}
	return false
}

// lookInFile opens a file and guesses if it is in old PDB format or
// in mmcif.
func lookInFile(fname string) (byte, error) {
	pdbWords := []string{"COMPND", "SOURCE", "REMARK", "SEQRES", "HETATM", "ATOM"}
	mmcifWords := []string{"data_", "entry.id", "loop_"}
	fp, err := os.Open(fname)
	if err != nil {
		return unk_fmt, err
	}
	defer fp.Close()

	rdr, e2 := zwrap.WrapMaybe(fp)
	if e2 != nil {
		return unk_fmt, errors.New("reading " + fname + " " + e2.Error())
	}

	const maxTestLines = 5000
	scnnr := bufio.NewScanner(bufio.NewReader(rdr))
	for i := 0; scnnr.Scan() && i < maxTestLines; i++ {
		s := scnnr.Text()
		for _, w := range mmcifWords {
			if comparefirst(s, w) {
				return mmcif_fmt, nil
			}
		}
		for _, w := range pdbWords {
			if comparefirst(s, w) {
				return old_fmt, nil
			}
		}
	}
	return unk_fmt, errors.New(fname + ": cannot recognise format")
}

// old_or_mmcif decides what format we will use.
// Maybe is uses the file name or maybe it peaks inside.
// We cannot use the function from filepath to get the file type,
// since it will return .gz if we feed it a.pdb.gz.
func oldOrMmcif(fname string) (byte, error) {
	s := filepath.Base(fname)
	var err error
	var i int
	if i = strings.IndexByte(s, '.'); i != -1 {
		s = s[i+1:] // change .ent to ent
		s = strings.ToLower(s)
		if strings.Contains(s, "pdb") || strings.Contains(s, "ent") {
			return old_fmt, nil
		} else if strings.Contains(s, "mmcif") || strings.Contains(s, "cif") {
			return mmcif_fmt, nil
		}
	}

	t := unk_fmt
	if t, err = lookInFile(fname); err != nil {
		return unk_fmt, err
	}
	return t, nil
}

// logWhere decide where to send output
func logWhere(outinfo string) (*log.Logger, error) {
	var iowriter io.Writer
	switch outinfo { // Decide where to send the logged output
	case "":
		iowriter = ioutil.Discard
	case "stdout":
		iowriter = os.Stdout
	default:
		var err error
		iowriter, err = os.OpenFile(outinfo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			return nil, err
		}
	}
	prefix := ""
	return log.New(iowriter, prefix, log.Lshortfile), nil
}

// ReadCoord takes a filename and attempts to read it as an
// mmcif file.
// Errors are returned the obvious way. During debugging, there can
// be a lot of output. This will be written to a file called outinfo.
// if outinfo is "", it will be trashed. If outinfo is "stdout", we
// write to standard output.
// Development ugliness - no longer handles http files.
func ReadCoord(fname string, srcType byte, outinfo string) ([]cmmn.Chain, error) {
	cmmnWanted := []string{
		"_entry.id",
		"_struct.title",
		"_pdbx_database_status.entry_id",
		"_pdbx_database_status.entry_id",
		"_entity_poly.pdbx_seq_one_letter_code",
		"refine.ls_d_res_high",
		"_reflns.d_resolution_high",
	}
	outlog, err := logWhere(outinfo)
	if err != nil {
		return nil, errors.New(err.Error() + " creating log file")
	}
	var rdr io.ReadCloser
	var typ byte
	if typ, err = oldOrMmcif(fname); err != nil {
		return nil, err
	} else if typ == old_fmt {
		return nil, errors.New("old format reading not written yet")
	} else {
		switch srcType {
		case cmmn.FileSrc:
			r1, e2 := os.Open(fname)
			if e2 != nil {
				return nil, e2
			}
			if rdr, e2 = zwrap.WrapMaybe(r1); err != nil {
				return nil, errors.New("reading " + fname + " " + e2.Error())
			}
		case cmmn.HTTPSrc:
			if rdr, err = getHTTP(fname, 2); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("programming bug")
		}
	}
	defer rdr.Close()
	mr := mmcif.NewMmcifReader(rdr)
	mr.AddItems(cmmnWanted)
	mr.AddTable([]string{"_entity_poly", "_database_2"})
	mr.SetChains([]string{})
	md, err := mr.DoFile()
	if err != nil {
		return nil, err
	}

	if md == nil {
		panic("programming bug")
	}

	valid, invalid := md.NAtomAllChainAll()
	outlog.Println(fname, valid, invalid)
	chains := md.GetChains()
	if chains == nil {
		return nil, errors.New("empty chains")
	}

	return chains, nil
}

// NatomsTot returns the total number of valid atoms and the number of invalid
// atoms in a set of chains. We cannot write it as a method, since it
// would have to be in the cmmn sub package.
func NatomsTot(chns []cmmn.Chain) (int, int) {
	var jValid, jInvalid int
	for _, c := range chns {
		cset := c.CoordSet
		for _, xyzS := range cset {
			for _, x := range xyzS {
				if x.Ok() {
					jValid++
				} else {
					jInvalid++
				}
			}
		}
	}
	return jValid, jInvalid
}

// NChain says how many chains we have in the structure
func NChain(chns []cmmn.Chain) int {
	return len(chns)
}

// ChainNames returns a slice of strings with the names of chains
// from a structure
func wasChainNames(chns []cmmn.Chain) (ret []string) {
	ret = make([]string, 0, 1)
	for _, k := range chns {
		ret = append(ret, k.ChainID)
	}
	return
}
