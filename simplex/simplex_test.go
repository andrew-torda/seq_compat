// 28 dec 2019

package simplex_test

import (
	"fmt"
	. "github.com/andrew-torda/goutil/simplex"
	"math"
	"math/rand"
	"testing"
)

// slicesDiffer returns true if two slices are not approximately the same.
// The definition of approximately is arbitrary. It is just enough
// for testing.
func slicesDiffer(x, y []float32) bool {
	const eps = 0.0001
	if len(x) != len(y) {
		panic("program bug slice lengths differ")
	}
	for i, v := range x {
		if math.Abs(float64(v-y[i])) > eps {
			return true
		}
	}
	return false
}

func costR(x []float32) (float32, error) {
	if x[0] > 4 {
		return 10000, nil
	}
	return (x[0] - 2) * (x[0] - 2), nil
}


const jnk float32 = 100

var tstPnt = [][]float32{
	{
		0.9, 1, jnk,
		1, 2, jnk,
		2, 1, jnk,
		3, 0, jnk,
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

var tstR1 = [][]float32{
	{3.1, 1, 100},
	{2.6, 1, 100},
	{-2.5, 1, 100},
}

// costbounds is used in the bounds tests
func costbounds(x []float32) (float32, error) {
	a := x[0] - 3
	return a * a, nil
}

// TestUpper tests upper bounds. The minimum is at 3, but are bound
// stops it going beyond 2.
func TestUpper(t *testing.T) {
	const ubound float32 = 2
	iniPrm := []float32{1, 95, 95, 95}
	s := NewSplxCtrl(costbounds, iniPrm, 300)
	s.Span([]float32{1, 3, 3, 3})
	s.Upper ([]float32{ubound, 100, 100, 100})
	res, err := s.Run(1)
	if err != nil {
		panic("badly written test in TestUpper")
	}
	if slicesDiffer (res.BestPrm[:1], []float32{ubound}) {
		t.Errorf ("TestUpper got %f for first element", res.BestPrm[:1])
	}
}
// TestLower for lower bounds
func TestLower(t *testing.T) {
	const lbound float32 = 4
	iniPrm := []float32{5, 110, 105, 105}
	s := NewSplxCtrl(costbounds, iniPrm, 300)
	s.Span([]float32{1, 3, 3, 3})
	s.Lower ([]float32{lbound, 100, 100, 100})
	res, err := s.Run(1)
	if err != nil {
		panic("badly written test in TestLower")
	}
	if slicesDiffer (res.BestPrm[:1], []float32{lbound}) {
		t.Errorf ("TestUpper got %f for first element", res.BestPrm[:1])
	}
}

// cost2 is a two parameter cost function
// (x-1)^2 + (y-5)^2
func cost2(x []float32) (float32, error) {
	a := (x[0] - 1)
	b := (x[1] - 5)
	return (a * a) + (b * b), nil
}

func nothing(interface{}) {}

func TestSimplexStruct(t *testing.T) {
	const a float32 = 5
	const b float32 = 5.1
	rr := rand.New(rand.NewSource(39499))
	noise50 := func(x float32) float32 { // noise50 takes a number
		fnoise := rr.Float32() - 0.5 // Between -1/2 and 1/2 // and adds a
		return fnoise*x + x          // random number within 1/2 of original value
	}
	correct := []float32{1, 5}
	for i := 0; i < 10; i++ { // Test with randomised starting points
		iniPrm := []float32{noise50(a), noise50(b)}
		s := NewSplxCtrl(cost2, iniPrm, 300)
		s.Scatter(0.4)
		res, err := s.Run(2)
		if err != nil {
			t.Errorf("prog bug testing")
		}
		if slicesDiffer(correct, res.BestPrm) {
			t.Errorf("simplex result got %v wanted %v", res.BestPrm, correct)
		}
	}
}

func costFlat(x []float32) (float32, error) {
	if x[0] < 1 || x[0] > 3 {
		return 1, nil
	}
	return float32(math.Abs(2.0 - float64(x[0]))), nil
}

// TestFlat tests a surface which is mostly flat.
// One dimension for the cost is enough.
func TestFlat(t *testing.T) {
	iniPrm := []float32{1.2, 99, 99, 99}
	s := NewSplxCtrl(costFlat, iniPrm, 200)
	s.Scatter(10)
	result, _ := s.Run(1)
	if slicesDiffer(result.BestPrm[:1], []float32{2}) {
		t.Errorf("Fail on flat surface")
	}
}

// costN puts minima a 1, 2, 3, ... in n dimensions.
func costN(x []float32) (float32, error) {
	var sum float32
	for i := 0; i < len(x); i++ {
		t := x[i] - float32(i+1)
		sum += t * t
	}
	return sum, nil
}

// TestNDim is for an n-dimensional simplex where n is something like seven.
func TestNDim(t *testing.T) {
	return
	iniPrm := []float32{10, 9, 8, 7, 6, 5, 4}
	s := NewSplxCtrl(costN, iniPrm, 500)
	s.Scatter(0.4)
	result, err := s.Run(1)
	if err != nil {
		t.Errorf("run failure in 7 dimensional test")
	}
	if slicesDiffer(result.BestPrm, []float32{1, 2, 3, 4, 5, 6, 7}) {
		t.Errorf("7 dimensional test Fail")
	}
}
