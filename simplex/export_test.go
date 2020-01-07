// 3 jan 2020
// The problem with a normal test of the simplex is that you will
// never see small mistakes. The method will probably converge, just
// inefficiently.
// I want to test some specific parts.

package simplex

import (
	"github.com/andrew-torda/goutil/matrix"
	"runtime"
	"sort"
)

type SWk struct{S sWk}
func (s *SplxCtrl) Onerun(sWk *sWk, splx splx) error {
	return s.onerun (sWk, splx)
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


func Amo1(splx splx, cost CostFun) ([]float32){
	var sWk sWk
	//	runtime.Breakpoint()
	nothing (runtime.Breakpoint)
	sWk.init(len(splx.Mat[0]), cost)
	sWk.setupFirstStep(splx)
	sort.Slice(sWk.rank, func(i, j int) bool {
		return sWk.y[sWk.rank[i]] > sWk.y[sWk.rank[j]]
	})
	sWk.centroid(splx)
	_, _ = amotry(splx, alpha, &sWk)
	//nothing(tRes)
	return splx.Mat[sWk.rank[0]] // highest vertex
}
