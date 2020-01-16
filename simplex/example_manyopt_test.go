// 13 jan 2020

package simplex_test

import (
	"fmt"
	"github.com/andrew-torda/goutil/simplex"
)

func cost(x []float32) (float32, error) {
	a := x[0] - 1
	b := x[1] - 4
	return a*a + b*b, nil
}

// Example_manyOpt exercises many options.
func Example_manyOpt() {
	iniPrm := []float32{1, 4}
	sctrl := simplex.NewSplxCtrl(cost, iniPrm, 400)
	if err := sctrl.Span([]float32{5, 6}); err != nil {
		panic("prog bug")
	}
	sctrl.Seed(7)   // an arbitrary number for random seed
	sctrl.Tol(1e-4) // not very tight tolerance
	sctrl.Lower([]float32{-3, -3})
	sctrl.Upper([]float32{5, 5})
	sctrl.IniType(simplex.IniPntSpread)
	res, _ := sctrl.Run(1)
	if slicesDiffer(res.BestPrm, []float32{1, 4}) {
		fmt.Println("wrong answer: ", res.BestPrm)
	} else {
		fmt.Println("OK")
	}
	// Output: OK
}
