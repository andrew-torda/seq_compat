// 28 dec 2019

package simplex_test

import (
	"errors"
	"fmt"
	. "github.com/andrew-torda/goutil/simplex"
	"testing"
)

// cost1 depends on one parameter, but it has two dimensions. We use it
// for testing some basic operations using helper functions in helper_test.go
func cost1(x []float32) (float32, error) {
	y := x[0] - 2
	y = y * y
	return y, nil
}

func costR(x []float32) (float32, error) {
	return (x[0] - 2) * (x[0] - 2), nil
}

var tstPnt = []float32{0.5, 1, 1, 2, 1, 0}

func TestReflect(t *testing.T) {

	npoint := 3
	nparam := 2
	iniPrm := []float32{0, 0}
	sctrl := NewSplxCtrl(cost2, iniPrm)
	xx := []float32{0.5, 1, 1, 2, 1, 0}
	splx := SplxFromSlice(npoint, nparam, xx)

	var n int
	for i := range splx.Mat { // Put array into simplex
		for j := range splx.Mat[i] {
			splx.Mat[i][j] = xx[n]
			n++
		}
	}
	fmt.Println("test splx\n", splx)
	nothing(splx)
	sctrl.Nopermute() // We do not want values re-ordered

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

func nothing(interface{}) {
}

func TestSimplexStruct(t *testing.T) {
	iniPrm := []float32{10, 20}
	s := NewSplxCtrl(cost2, iniPrm)
	s.Run(100, 3)
}
