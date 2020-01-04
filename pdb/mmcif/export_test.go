package mmcif
import "github.com/andrew-torda/goutil/pdb/cmmn"

// Export some internal functions for testing

func (s *cmmtScanner) Cbytes() []byte   { return s.cbytes() }
func (s *cmmtScanner) Cscan() (ok bool) { return s.cscan() }

var NewCmmtScanner = newCmmtScanner
var SplitCifLine = splitCifLine
var Fields = fields

func (md *MmcifData) NAtomType(s string) int { return md.nAtomType(s) }

func (md *MmcifData) GetXyz(mdlNum int16, chn string, at string) (cmmn.XyzSl, error) {
	return md.getXyz(mdlNum, chn, at)
}

type BSlice = bSlice

