// 3 jan 2020

package simplex

// rot rotates the points in the simplex. It should not make
// any different to the results.
func (splx splx) Rot() {
	a := make([]float32, len(splx.Mat[0]))
	copy(a, splx.Mat[0])
	for i, v := range splx.Mat[1:] {
		copy(splx.Mat[i], v)
	}
	copy(splx.Mat[len(splx.Mat)-1], a)
}
