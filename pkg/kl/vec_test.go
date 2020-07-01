// 1 Jul 2020

// vector checks

package kl_test

import (
	. "github.com/andrew-torda/seq_compat/pkg/kl"
	"math"
	"testing"
)

func approxequal(x, y float32) bool {
	const eps float32 = 0.0001
	d := x - y
	if d < 0 {
		d = -d
	}
	if d < eps {
		return true
	}
	return false
}

func sqrt(x float32) float32 { return float32(math.Sqrt(float64(x))) }

func TestInnerCosProd(t *testing.T) {
	sqrt2 := float32(math.Sqrt (2))
	var v_scr []float32
	var space [64]float32 // no need for a call to make()
	scale := func(v []float32, x float32) []float32 {
		v_scr = space[:len(v)]
		for i := range v {
			v_scr[i] = v[i] * x
		}
		return v_scr
	}
	type f32 []float32
	pairs := []struct {
		v1  f32
		v2  f32
		res float32
	}{
		{f32{1, 0}, f32{0, 1}, 0},
		{f32{1, 1, 0, 0}, f32{0, 0, 1, 1}, 0},
		{f32{1, 2, 3}, f32{1, 2, 3}, 1},
		{f32{0, 1}, f32{1, 1}, 1 / sqrt2},
		{f32{0, 1, 0}, f32{1, 1, 0}, 1 / sqrt2},
		{f32{0, 0, 1}, f32{0, 1, 1}, 1 / sqrt2},
		{f32{0, 0, 0, 0, 1}, f32{0, 0, 0, 1, 1}, 1 / sqrt2},
	}
	// Multiplying a vector by itself gives 1. Scaling a vector by a positive
	// number still gives one. Flipping the sign should give the opposite of the
	// original result.
	scaleVal := f32{1, 13, 0.15}
	for _, pair := range pairs {
		for _, x := range scaleVal {
			r := InnerCosProd(pair.v1, scale(pair.v1, x))
			s := InnerCosProd(scale(pair.v1, x), pair.v1)
			tt := InnerCosProd(pair.v1, scale(pair.v1, -x))
			if r != s {
				t.Fatalf("Self multiply argument order r %f s %f, vec %v", r, s, pair.v1)
			}
			if r != 1 {
				t.Fatal("Self multiply scale", x, "wanted 1, got ", r, pair.v1, v_scr)
			}
			if !approxequal(tt, -1) {
				t.Fatal("Sign swap gave", tt, "for", pair.v1)
			}
		}
	}

	for _, pair := range pairs {
		for _, x := range scaleVal {
			r := InnerCosProd(scale(pair.v1, x), pair.v2)
			s := InnerCosProd(scale(pair.v1, -x), pair.v2)
			if !approxequal(r, pair.res) {
				t.Fatalf("Got %f wanted %f, pair %v, %v", r, pair.res, scale(pair.v1, x), pair.v2)
			}
			if !approxequal(r, -s) {
				t.Fatal("swapped signed r != -s", r, s)
			}
		}
	}
}
