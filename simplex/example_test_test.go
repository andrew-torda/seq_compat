// This is a test file, but of a test function, hence the funny name.

// You have a fitting exercise like y = ax^2 + bx + c based on n points.
// You do the fitting on 90% of the data, but reserve 10% for testing.

package simplex_test

import (
	"fmt"
	"github.com/andrew-torda/goutil/simplex"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"runtime"
)

func quadratic(a, b, c float32, x float32) (r float32) {
	r = a * x * x
	r += b * x
	r += c
	return r
}

func fitQuality(a, b, c float32, x, y []float32) float32 {
	var sum float32
	for i, v := range x {
		tmp := quadratic(a, b, c, v)
		diff := y[i] - tmp
		diff *= diff
		sum += diff
	}
	return sum
}

func noiseR(x, r float32) float32 { // add r amount of noise to x
	fnoise := rand.Float32() - 0.5 // a number from -1/2 to 1/2
	return fnoise*x + x            // original +- r/2*x
}

// makeCostTest returns a main cost function and a function for testing.
// Make an array with data from a simple function and split into
// 90% and 10% pieces. The main cost function acts on 90% of the data.
// The test function on 10%.
func makeCostTest() (simplex.CostFun, simplex.CostFun) {
	const (
		ndata       int     = 200
		xmin        float32 = -30
		xmax        float32 = 30
		noise       float32 = 0.02 // 2 % noise
		a           float32 = 2    // for y = ax^2 + b + c
		b           float32 = 5
		c           float32 = 5
		fractesting float64 = 0.2
	)
	xdata := make([]float32, ndata)
	ydata := make([]float32, ndata)
	inc := (xmax - xmin) / float32(ndata-1)
	xdata[0] = xmin
	for i := 1; i < ndata; i++ {
		xdata[i] = xdata[i-1] + inc
	}

	rand.Shuffle(len(xdata), func(i, j int) {
		xdata[i], xdata[j] = xdata[j], xdata[i]
	})

	for i, v := range xdata {
		ydata[i] = quadratic(a, b, c, v)
		ydata[i] = noiseR(ydata[i], noise)
	}
	// Make a main slice and test data slice
	ix := int32(math.Round((1 - fractesting) * float64(ndata)))
	mainXdata := xdata[:ix]
	mainYdata := ydata[:ix]
	testXdata := xdata[ix:]
	testYdata := ydata[ix:]
	if (len(mainXdata) + len(testXdata)) != ndata {
		panic("programming bug")
	}
	// These are what we will return - wrappers around fitQuality with
	// different data sets.
	costMain := func(x []float32) (float32, error) {
		return fitQuality(x[0], x[1], x[2], mainXdata, mainYdata), nil
	}
	costTest := func(x []float32) (float32, error) {
		return fitQuality(x[0], x[1], x[2], testXdata, testYdata), nil
	}

	return costMain, costTest
}

// tmpfname gives me a safe name for a temporary file.
// We call the library to get a file handle, then get its name, then close it.
func tmpfname() (string, error) {
	fp, err := ioutil.TempFile(".", "tmp")
	if err != nil {
		return "", err
	}
	name := fp.Name()
	fp.Close()
	return name, nil
}

// sz returns size of a file
func sz(name string) (sz int64) {
	finfo, err := os.Stat(name)
	if err != nil {
		panic("sz should not fail in testing")
	}
	return finfo.Size()
}

func Example_CostTest() {
	costfun, testfun := makeCostTest()
	iniPrm := []float32{2, 5, -25}
	sctrl := simplex.NewSplxCtrl(costfun, iniPrm, 500)
	nothing(runtime.Breakpoint)
	sctrl.Span([]float32{.5, .5, .5})
	testFname, _ := tmpfname()
	historyFname, _ := tmpfname()

	sctrl.Testfun(testfun, testFname)
	defer os.Remove(testFname)
	sctrl.HstryFname(historyFname)
	defer os.Remove(historyFname)
	r, err := sctrl.Run(2)
	if err != nil {
		fmt.Println("Example_CostTest", err)
		return
	}
    if r.StopReason != simplex.Converged {
		fmt.Println ("CostTest really should converge")
	}
	if sz(testFname) < 100 {
		fmt.Println("output in test data too small")
		return
	}
	if sz(historyFname) < 100 {
		fmt.Println("output in history file too small")
		return
	}
	fmt.Println("OK")
	// Output: OK
}
