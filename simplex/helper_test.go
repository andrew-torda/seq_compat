// 3 jan 2020
// The problem with a normal test of the simplex is that you will
// never see small mistakes. The method will probably converge, just
// inefficiently.
// I want to test some specific parts.

package simplex

func newSWk (npnt, ndim int) (sWk){
	var sWk sWk
	sWk.y = make([]float32, npnt)
	sWk.rank = make([]int, npnt)
	sWk.cntrd = make([]float32, ndim)
	sWk.ptrial = make([]float32, ndim)
	return sWk
}
func Amo1 (splx splx) {
	sWk := newSWk (3, 2)
	tRes, _ := amotry (splx, alpha, sWk)
	nothing (tRes)
}
