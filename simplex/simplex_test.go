// 28 dec 2019

package simplex_test

import (
	"fmt"
	. "github.com/andrew-torda/goutil/simplex"
	"math"
	"math/rand"
	"os"
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
	s.Upper([]float32{ubound, 100, 100, 100})
	res, err := s.Run(1)
	if err != nil {
		panic("badly written test in TestUpper")
	}
	if slicesDiffer(res.BestPrm[:1], []float32{ubound}) {
		t.Errorf("TestUpper got %f for first element", res.BestPrm[:1])
	}
}

// TestLower for lower bounds
func TestLower(t *testing.T) {
	const lbound float32 = 4
	iniPrm := []float32{5, 110, 105, 105}
	s := NewSplxCtrl(costbounds, iniPrm, 300)
	s.Span([]float32{1, 3, 3, 3})
	s.Lower([]float32{lbound, 100, 100, 100})
	res, err := s.Run(1)
	if err != nil {
		panic("badly written test in TestLower")
	}
	if slicesDiffer(res.BestPrm[:1], []float32{lbound}) {
		t.Errorf("TestUpper got %f for first element", res.BestPrm[:1])
	}
}

// cost2 is a two parameter cost function
// (x-1)^2 + (y-5)^2
func cost2(x []float32) (float32, error) {
	a := (x[0] - 1)
	b := (x[1] - 5)
	return (a * a) + (b * b), nil
}

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
		s := NewSplxCtrl(cost2, iniPrm, 200)
		s.Scatter(0.4)
		res, err := s.Run(2)
		if err != nil {
			t.Errorf("prog bug testing")
		}
		if slicesDiffer(correct, res.BestPrm) {
			t.Errorf("simplex got %v wanted %v starting from %v repetition %v",
				res.BestPrm, correct, iniPrm, i)
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
	iniPrm := []float32{10, 9, 8, 7, 6, 5, 4}
	s := NewSplxCtrl(costN, iniPrm, 800)
	s.Scatter(0.4)
	result, err := s.Run(1)
	if err != nil {
		t.Errorf("run failure in 7 dimensional test")
	}
	if slicesDiffer(result.BestPrm, []float32{1, 2, 3, 4, 5, 6, 7}) {
		s := fmt.Sprintln(result.BestPrm)
		t.Errorf("7 dimensional test Fail reached " + s)
	}
}

// TestSetupErr is to make sure that we really do flag an error when we set up with
// a slice of the wrong dimension.
func TestSetupErr(t *testing.T) {
	s := NewSplxCtrl(costN, []float32{1, 2, 3}, 100)
	err := s.Span([]float32{1, 2, 3, 4})
	if err == nil {
		t.Errorf("slice check failed")
	}
}

// TestCvgFlat invokes the code to keep contracting when we hit a flat region
// of the cost function.
// Check the second convergence criterion, which is helpful on a flat surface.
// This is no longer a real test. I checked the values in a debugger, but you have
// to look at the entries in the simplex and these are hidden from the outside world.
func TestCvgFlat(t *testing.T) {
	iniPrm := []float32{2, 1, 1}
	costflat := func(x []float32) (float32, error) {
		if x[0] < 1 {
			return 1 - x[0], nil
		}
		if x[0] > 3 {
			return x[0] - 3, nil
		}
		return 0, nil
	}
	s := NewSplxCtrl(costflat, iniPrm, 300)
	s.Span([]float32{2.1, 1, 1})
	s.Ptol([]float32{0.01, 0.01, 0.01})
	r, _ := s.Run(1)
	if r.BestPrm[0] < 1 || r.BestPrm[0] > 3 {
		t.Errorf("TestCvgFlat broke")
	}
}

//costerr is a silly cost function
func costerr([]float32) (float32, error) {
	return 1, fmt.Errorf("Artificial error to check code")
}

// TestCostErr checks if errors really get passed back from the cost function
func TestCostErr(t *testing.T) {
	iniPrm := []float32{1, 1, 1}
	s := NewSplxCtrl(costerr, iniPrm, 100)
	_, err := s.Run(2)
	if err == nil {
		t.Errorf("Should have passed error back to caller")
	}
}

func costlinear(x []float32) (r float32, err error) {
	if x[0] < 0 {
		r = -2.0 * x[0]
	} else {
		r = x[0]
	}
	return r, nil
}

func TestA1(t *testing.T) {
	iniPrm := []float32{-1, 99}
	s := NewSplxCtrl(costlinear, iniPrm, 100)
	s.Span([]float32{2, 1e-7})
	r, _ := s.Run(1)
	if slicesDiffer(r.BestPrm[:1], []float32{0}) {
		t.Errorf("testA1 got %v not %v", r.BestPrm, "zeroes")
	}
}

func multilinear(x []float32) (r float32, err error) {
	var sum float32 = 0
	for _, v := range x {
		if v < 0 {
			sum += -2.0 * v
		} else {
			sum += v
		}
	}
	return sum, nil
}

func TestA2(t *testing.T) {
	const historyFile string = "historyFile"
	iniPrm := []float32{-1, 5, -10}
	s := NewSplxCtrl(multilinear, iniPrm, 100)
	s.Span([]float32{3, 3, 3})
	s.HstryFname(historyFile)
	s.IniType(IniClassic)
	s.Tol(0.001)
	r, _ := s.Run(2)
	zeroes := make([]float32, len(iniPrm))
	if slicesDiffer(r.BestPrm, zeroes) {
		t.Errorf("testA2 got %v not %v", r.BestPrm, "zeroes")
	}
	os.Remove(historyFile)
}

func innerlinear(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// TestDimensions has a look if everything works, even when the important dimension
// changes. Don't care about dimensions so much, but about indexing.
func TestDimensions(t *testing.T) {
	const (
		ndim        int     = 5
		spanconst   float32 = 5
		iniPrmRange         = 20
		eps         float64 = 0.001
	)
	span := make([]float32, ndim)
	for j := range span {
		span[j] = spanconst
	}
	iniPrm := make([]float32, ndim)
	for i := 0; i < ndim; i++ {
		for i := range iniPrm {
			x := rand.Float32()
			x = x - 0.5
			iniPrm[i] = x * iniPrmRange
		}
		cost := func(x []float32) (float32, error) {
			return innerlinear(x[i]), nil
		}
		s := NewSplxCtrl(cost, iniPrm, 100)
		s.Span(span)
		r, _ := s.Run(2)
		if math.Abs(float64(r.BestPrm[i])) > eps {
			t.Errorf("TestDimensions dimension %d should be near 0.0, got %f",
				i, r.BestPrm[i])
		}
	}
}
