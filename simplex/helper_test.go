// 3 jan 2020
// The problem with a normal test of the simplex is that you will
// never see small mistakes. The method will probably converge, just
// inefficiently.
// I want to test some specific parts.

package simplex

import (
	"github.com/andrew-torda/goutil/matrix"
)

func SplxFromSlice(npoint, nparam int, x []float32) splx {
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

var NewSwk = newSWk

func Amo1(splx splx) {
	sWk := newSWk(3, 2)
	tRes, _ := amotry(splx, alpha, sWk)
	nothing(tRes)
}
