//

package geom_test

import (
	"math"
	. "github.com/goutil/pdb/calpha/geom"
	. "github.com/goutil/pdb/cmmn"
	"testing"
)

var disttests = []struct {
	name string
	x1   Xyz
	x2   Xyz
	res  float32
	e    bool
}{
	{"3.8 ", Xyz{3.80, 0.00, 0}, Xyz{0, 0, 0}, Brokendist, false},
	{"onex", Xyz{0.00, 0.00, 0}, Xyz{1, 0, 0}, Brokendist, true},
	{"333 ", Xyz{3.00, 3.00, 3}, Xyz{1, 0, 0}, Brokendist, false},
	{"333 ", Xyz{1.95, 1.95, 3}, Xyz{0, 0, 0}, Brokendist, false},
	{"5.0 ", Xyz{5.00, 5.00, 5}, Xyz{1, 0, 0}, Brokendist, true},
}

// permuteXyz rotates x, y znd z for tests whose answers should not change
// when we move the axes around.
func permuteXyz(x Xyz) Xyz {
	x.X, x.Y, x.Z = x.Y, x.Z, x.X
	return x
}

func TestXyzDist(t *testing.T) {
	for _, test := range disttests {
		var dist1, dist2, dist3, dist4 float32
		var e1, e2, e3, e4 error
		x1 := test.x1
		x2 := test.x2
		dist1, e1 = XyzDist(x1, x2)
		dist2, e2 = XyzDist(x2, x1)
		x1, x2 = permuteXyz(x1), permuteXyz(x2)
		dist3, e3 = XyzDist(x1, x2)
		x1, x2 = permuteXyz(x1), permuteXyz(x2)
		dist4, e4 = XyzDist(x1, x2)
		if dist1 != dist2 || dist1 != dist3 || dist1 != dist4 {
			t.Errorf("test %s. Did not get identical results, %f %f %f %f",
				test.name, dist1, dist2, dist3, dist4)
		}
		if e1 != e2 || e1 != e3 || e1 != e4 {
			t.Errorf("test %s, did not get the same error state", test.name)
		}
		if test.e {
			if e1 == nil {
				t.Errorf("test %s expected error did not get one", test.name)
			}
		} else {
			if e1 != nil {
				t.Errorf("test %s got unexpected error %s", test.name, e1)
			}
		}
	}
}

// notApproxEqual returns true if x and y are not approximately equal.
func notApproxEqual(x, y float32) bool {
	diff := x - y
	if diff < 0 {
		diff = -diff
	}
	if math.IsNaN(float64(diff)) {
		return true
	}
	if diff > 0.00001 {
		return true
	}
	return false
}

var angletests = []struct {
	x1, x2, x3 Xyz
	res        float32
}{
	{Xyz{+1, 0, 0}, Xyz{0, 0, 0}, Xyz{0.9999, 0, 0}, 0},
	{Xyz{-0, 1, 0}, Xyz{0, 0, 0}, Xyz{1.0000, 0, 0}, math.Pi / 2},
	{Xyz{-1, 0, 0}, Xyz{0, 0, 0}, Xyz{1.0000, 0, 0}, math.Pi},
	{Xyz{+0, 1, 0}, Xyz{0, 0, 0}, Xyz{1.0000, 0, 0}, math.Pi / 2},
	{Xyz{+0, 1, 0}, Xyz{0, 0, 0}, Xyz{0.1000, 0, 0}, math.Pi / 2},
	{Xyz{+0, 1, 0}, Xyz{0, 0, 0}, Xyz{9.9000, 0, 0}, math.Pi / 2},
	{Xyz{-1, 0, 0}, Xyz{0, 0, 0}, Xyz{1.0000, 1, 0}, math.Pi * 3 / 4},
	{Xyz{-1, 0, 0}, Xyz{0, 0, 0}, Xyz{9.9, 9.9, 0}, math.Pi * 3 / 4},
}

func TestXyzAngle(t *testing.T) {
	for _, test := range angletests {
		x1, x2, x3 := test.x1, test.x2, test.x3
		if a, err := XyzAngle(x1, x2, x3); err != nil {
			t.Errorf("%v error with %v %v %v", err, x1, x2, x3)
		} else {
			if notApproxEqual(a, test.res) {
				t.Errorf("TestXyzAngle got %f wanted %f, %v, %v, %v",
					a, test.res, x1, x2, x3)
			}
		}
		x1, x2, x3 = permuteXyz(x1), permuteXyz(x2), permuteXyz(x3)
		if a, err := XyzAngle(x1, x2, x3); err != nil {
			t.Errorf("%v error with %v %v %v", err, x1, x2, x3)
		} else {
			if notApproxEqual(a, test.res) {
				t.Errorf("TestXyzAngle got %f wanted %f, %v, %v, %v",
					a, test.res, x1, x2, x3)
			}
		}

		x1, x2, x3 = permuteXyz(x1), permuteXyz(x2), permuteXyz(x3)
		if a, err := XyzAngle(x1, x2, x3); err != nil {
			t.Errorf("%v error with %v %v %v", err, x1, x2, x3)
		} else {
			if notApproxEqual(a, test.res) {
				t.Errorf("TestXyzAngle got %f wanted %f, %v, %v, %v",
					a, test.res, x1, x2, x3)
			}
		}

	}
}

var dhdrltests = []struct {
	x1, x2, x3, x4 Xyz
	res            float32
}{
	{Xyz{0, 1, 0}, Xyz{1, 0, 0}, Xyz{2, 0, 0}, Xyz{3,  1, 0},  0},
	{Xyz{0, 1, 0}, Xyz{1, 0, 0}, Xyz{2, 0, 0}, Xyz{3,  1, 1e-5},  0},
	{Xyz{0, 1, 0}, Xyz{1, 0, 0}, Xyz{2, 0, 0}, Xyz{3, -1, 0},  math.Pi},
	{Xyz{0, 1, 0}, Xyz{1, 0, 0}, Xyz{2, 0, 0}, Xyz{3,  0, 1}, -math.Pi/2},
	{Xyz{0, 1, 0}, Xyz{1, 0, 0}, Xyz{2, 0, 0}, Xyz{3,  0,-1},  math.Pi/2},
	{Xyz{0, 1, 0}, Xyz{1, 0, 0}, Xyz{2, 0, 0}, Xyz{3,  1,-1},  math.Pi/4},
	{Xyz{0, 1, 0}, Xyz{1, 0, 0}, Xyz{2, 0, 0}, Xyz{3, -1,-1},  math.Pi * (3.0/4.0)},
}
// To add
// Move the inner loop into another function and then we can...
//  1. reverse the order of points and repeat
//  2. increase distance between middle points and repeat
//  3. shift all points by a unit along an axis and repeat
func TestXyzDhdrl(t *testing.T) {
	for _, test := range dhdrltests {
		x1, x2, x3, x4 := test.x1, test.x2, test.x3, test.x4
		const emsg = "error with %v %v %v %v wanted: %.3g got: %.3g"
		for i := 0; i < 3; i++ {
			a := XyzDhdrl(x1, x2, x3, x4)
			if notApproxEqual(a, test.res) {
				t.Errorf(emsg, x1, x2, x3, x4, test.res, a)
			}
			x1, x2, x3, x4 = permuteXyz(x1), permuteXyz(x2), permuteXyz(x3), permuteXyz(x4)
		}
	}
}
