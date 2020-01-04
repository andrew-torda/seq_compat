package mmcif

import (
	"pdb/cmmn"
)

// GetChains takes the result of the mmcif reader and returns a slice
// of the simpler structure, Chain
func (md *MmcifData) GetChains () ([]cmmn.Chain) {
	ret := make ([]cmmn.Chain, 0)
	for chainid, onechain := range md.Allcoord {
		ch := new (cmmn.Chain)
		ch.ChainID = string(chainid)
		ch.MdlNum = 0
		ch.NumLbl = onechain.numLbl
		ch.InsCode = onechain.insCode
		ch.CoordSet = make (cmmn.CoordSet)
		for atname, xyzsl := range onechain.coords[0] {
			ch.CoordSet[string(atname)] = xyzsl
		}
		ret = append (ret, *ch)
	}
	return ret
}
