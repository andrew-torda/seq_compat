// 28 dec 2019

package simplex_test

import (
	"errors"
	. "github.com/andrew-torda/goutil/simplex"
	"math"
	"testing"
)

// slicesDiffer returns true if two slices are not approximately the same.
// The definition of approximately is arbitrary. It is just enough
// for testing.
func slicesDiffer(x, y []float32) bool {
	const eps = 0.00001
	for i, v := range x {
		if math.Abs(float64(v - y[i])) > eps {
			return true
		}
	}
	return false
}

// cost1 depends on one parameter, but it has two dimensions. We use it
// for testing some basic operations using helper functions in helper_test.go
func cost1(x []float32) (float32, error) {
	y := x[0] - 2
	y = y * y
	return y, nil
}

// rotSl shuffles the elements of a slice
func rotSl(f []float32) []float32 {
	return (append(f[1:], f[0]))
}

func costR(x []float32) (float32, error) {
	if (x[0] > 4) { return 10000, nil }
	return (x[0] - 2) * (x[0] - 2), nil }

const shifter float32 = 0.5
const jnk float32 = 100
var tstPnt = [][]float32 {
	{
		0.9, 1, jnk,
		2-(2*shifter), 2, jnk,
		2, 1, jnk,
		2+(2*shifter), 0, jnk,
	},
	{
		0.4, 1, jnk,
		0.5, 2, jnk,
		1.5, 1, jnk,
		2.5, 0, jnk,
	},
	{
		-6.5, 1, jnk,
		0.5, 2, jnk,
		1.5, 1, jnk,
		2.5, 0, jnk,
	},
}

var tstR1 = [][]float32 {
	{3.1, 1, 100},
	{2.6, 1,100},
	{-2.5, 1, 100},
}
// TestR2 is for checking reflections, extensions and 1D contraction.
func TestR2(t *testing.T) {
	iniPrm := []float32{0, 0, 0}
	sctrl := NewSplxCtrl(cost2, iniPrm)
	sctrl.Maxstep(1)
	sctrl.Nopermute() // Do not want values re-ordered
	ndim := len(iniPrm)
	for i := range tstPnt {
		for n := 0; n <= ndim; n++ {
			var sWk SWk
			sWk.Init(ndim, costR)
			splx := SplxFromSlice(ndim, tstPnt[i])
			sctrl.Onerun(&sWk.S, splx)
			hiPnt := splx.Mat[0]
			if slicesDiffer(hiPnt, tstR1[i]) {
				t.Errorf("reflect test high point high %v, expected %v ", hiPnt, tstR1[i])
			}
			splx.Rot()
		}
	}
}

// cost2 is a two parameter cost function
// (x-1)^2 + (y-5)^2
func cost2(x []float32) (float32, error) {
	if len(x) != 2 {
		return 0, errors.New("bad arg to cost2")
	}
	a := (x[0] - 1)
	b := (x[1] - 5)
	return (a * a) + (b * b), nil
}

func nothing(interface{}) {}

func TestSimplexStruct(t *testing.T) {
	iniPrm := []float32{10, 20}
	s := NewSplxCtrl(cost2, iniPrm)
	s.Run(100, 3)
}
