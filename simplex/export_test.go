// 3 jan 2020
// The problem with a normal test of the simplex is that you will
// never see small mistakes. The method will probably converge, just
// inefficiently.
// I want to test some specific parts.

package simplex

import (
	"github.com/andrew-torda/goutil/matrix"
)

type SWk struct{S sWk}
func (s *SplxCtrl) Onerun(sWk *sWk, splx splx) (uint8, error) {
	return s.onerun (sWk)
}

func (s *SWk) Init(ndim int, cost CostFun) {
	s.S.init (ndim, cost)
}

// rot rotates the points in the simplex. It should not make
// any different to the results.
func (splx splx) Rot () {
	a := make ([]float32, len(splx.Mat[0]))
	copy (a, splx.Mat[0])
	for i, v := range splx.Mat[1:] {
			copy (splx.Mat[i], v)
	}
	copy (splx.Mat[len(splx.Mat) - 1], a)
}


func SplxFromSlice(nparam int, x []float32) splx {
	npoint := nparam + 1
	if npoint*nparam != len(x) {
		panic("Should not happen in testing")
	}
	splx := splx{matrix.NewFMatrix2d(npoint, nparam)}
	n := 0
	for i := range splx.Mat { // Put array into simplex
		for j := range splx.Mat[i] {
			splx.Mat[i][j] = x[n]
			n++
		}
	}
	return splx
}
