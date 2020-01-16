// Classic Rosenbrock function
package simplex_test

import (
	"fmt"
	. "github.com/andrew-torda/goutil/simplex"
)

// The Rosenbrock function is a classic example
// f(x, y) = (a-x)^2  + b (y -x^2)^2
// with a minimum at x,y = a, a^2
// The method is actually rather fragile. It does not work with some
// starting points and one start. It is much more tolerant if you give
// it two starts (s.Run(2)).
func ExampleRosenbrock() {
	const a float32 = 2
	const b float32 = 100
	costRosenbrock := func(x []float32) (float32, error) {
		p := a - x[0]
		q := x[1] - x[0]*x[0]
		return p*p + b*q*q, nil
	}

	iniPrm := []float32{-1, 9}
	s := NewSplxCtrl(costRosenbrock, iniPrm, 500)
	s.Span([]float32{3, 3})
	r, _ := s.Run(2)
	if slicesDiffer(r.BestPrm, []float32{a, a * a}) {
		fmt.Println("answer wrong", r.BestPrm)
	} else {
		fmt.Println("ok at 1, 1")
	}
	// Output: ok at 1, 1
}
