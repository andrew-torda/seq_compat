// Package seq provides a simplex (Nelder and Meade) optimizer.
// It follows the original paper
//    J.A. Nelder, R. Mead (1965), Comp. J., 7, 308-313
// but some of the nomenclature comes from
//    J.C. Lagarias, J.A. Reeds, M.H. Wright, and P.E. Wright (1998),
//    SIAM J. Optim, 9, 112-147.
// and some from ..
//  Press, W.H., Teukolsky, S.A., Vetterling, W.T., Flannery, B.P.,
//  Numerical Recipes in C., Cambridge University Press, 1992
// The structure with the amotry() function comes from numerical recipes,
// but the formulae for moving the highest point around are taken from
// the primary references.
// It has a couple of frills.
//  1. There are two ways to initialise. You can use the classic version as in
//     numerical recipes. Works well on my very artificial examples which have
//     some funny symmetries.
//     Alternatively, you can the initial points in each dimension spread evenly
//     over the allowed values and surrounding the initial parameter values.
//  2. It allows a vector of minimum and maximum values. It will reject
//     moves if they go beyond these boundaries. This is done by wrapping
//     the cost function.
// Could be improved
//
//  * we call a full sort after each cycle. This is not necessary. One
//    only needs a list with the highest, next-highest and best points.
//  * We re-calculate the centroid on each cycle. This is necessary, but
//    the version in numerical recipes does it by removing the old best
//    value and adding in the new one. In a few dimensions, it makes no
//    difference. In many dimensions, one could argue the method in numerical
//    recipes will be faster.
package simplex

import (
	"fmt"
	"github.com/andrew-torda/goutil/matrix"
	"math"
	"math/rand"
	"os"
	"sort"
)

const (
	debugtillIbleed bool = false // Writes lots of stuff to a file called dbg
)

type iniType uint8

const ( // These two have to be exported, so we can say simplex.IniClassic
	IniClassic iniType = iota
	IniPntSpread
)
const (
	Converged uint8 = iota // Stopped after converging
	MaxSteps               // Stopped after max steps
	Bust                   // Stopped after an error
)

const (
	randSeed         = 1637    // default seed for permuting
	alpha    float32 = 1       // Literature values for
	beta     float32 = -1 / 2. // moving the worst vertex in
	gamma    float32 = -2      // the simplex
)

// let's add
//  - filename for min and max values
//  - trajectory slices - for storing and printing out our trajectory.
// cost is the cost function, but testfun lets us invoke something
// (probably the same basic function), but perhaps on a test data
// set.
//

// CostFun is the function to be optimised. It must accept a slice of
// floats and return a single float as the answer as well as an
// error. If one needs more, such as fixed parameters, use an
// appropriate wrapper.
type CostFun func(x []float32) (float32, error)

// SplxCtrl is the structure for controlling how the simplex
// operates. All the parameters can be set via the methods that act on
// an SplxCtrl.
type SplxCtrl struct {
	lower      []float32 // bounds on the parameters
	upper      []float32
	iniPrm     []float32 // initial parameter guess
	span       []float32 // initial range for each parameter
	pTol       []float32 // optional tolerance for each parameter.
	hstryFname string    // Filename to which history will be written
	testFname  string    // Filename to which test function results go
	seed       int       // Seed for random number generator
	cost       CostFun   // the function to be optimised
	testfun    CostFun   // A function called each time simplex improve
	scatter    float32   // Spread around initial simplex points
	tol        float32   // Tolerance criterion
	maxstep    int32     // cycles of the simplex
	verbosity  int8      // for debugging
	noPermute  bool      // turn off permuting of simplex at setup
	iniType    iniType   // do classic initialisation of simplex points
}

// NewSplxCtrl returns a structure for controlling a simplex
// optimizer. After getting the return value, you call the various
// setting functions to specify parameters.
// cost is the cost function.
func NewSplxCtrl(cost CostFun, iniPrm []float32, maxstep int32) *SplxCtrl {
	s := new(SplxCtrl)
	s.cost = cost
	s.iniPrm = iniPrm
	s.maxstep = maxstep
	s.tol = 1e-5
	s.seed = randSeed
	s.scatter = 0.2 // 20 % scatter means original +/- 10 %
	rand.Seed(int64(s.seed))
	s.iniType = IniClassic
	const forcesquash = false
	return s
}

// Result is the structure used to pass results back to the caller
//
// We are told the number of cycles used, the final best parameters
// (BestPrm) and the reason for stopping.
type Result struct {
	Ncycle     int32     // How many times did we cycle
	BestPrm    []float32 // best parameters found
	StopReason uint8     // Reason why calculation stopped
}

// checklen is used in the set functions which take a slice (usually of float32's).
func checklen(name string, got []float32, want []float32) error {
	ngot := len(got)
	nwant := len(want)
	if ngot != nwant {
		e := fmt.Errorf("check param \"name\". Len mismatch. Got %d. Want %d",
			ngot, nwant)
		return e
	}
	return nil
}

// Testfun adds a testfunction to the control structure. Every time we
// improve the best point in the simplex, we invoke this
// function. This lets us do bootstrapping.
// We are given the name of a file to which results will be written.
func (s *SplxCtrl) Testfun(c CostFun, fname string) {
	s.testfun = c
	s.testFname = fname
}

// Initype sets the initialisation to either classic or spread points. Call it
// as s.Initype (simplex.IniClassic) or s.Initype (simplex.IniPntSpread)
func (s *SplxCtrl) IniType(i iniType) { s.iniType = i }

// verbosity
func (s *SplxCtrl) Verbosity(i int8) { s.verbosity = i }

// HstryFname is the filename to which history of the
// optimization will be written.
func (s *SplxCtrl) HstryFname(name string) { s.hstryFname = name }

// Scatter sets the scatter around the starting point in each dimension
func (s *SplxCtrl) Scatter(f float32) { s.scatter = f }

// Seed sets the random number seed used in permuting indices
func (s *SplxCtrl) Seed(i int) { s.seed = i }

// Nopermute turns off the permuting of initial values in the simplex.
// This is only useful in debugging
func (s *SplxCtrl) Nopermute() { s.noPermute = true }

// Tol sets the convergence tolerance.
func (s *SplxCtrl) Tol(f float32) { s.tol = f }

// Span sets the initial ranges for parameters
func (s *SplxCtrl) Span(x []float32) error {
	if err := checklen("span", x, s.iniPrm); err != nil {
		return err
	}
	s.span = x
	return nil
}

// Ptol sets a tolerance for each parameter individually. It is optional.
// If the tol criterion is met, then we can ask in each dimension if
// the x_min - x_max is smaller than some value.
// This is useful for noisy cost functions and will make the simplex keep
// contracting a bit more.
func (s *SplxCtrl) Ptol(pTol []float32) error {
	if err := checklen("pTol", pTol, s.iniPrm); err != nil {
		return err
	}
	s.pTol = pTol
	s.pTol = pTol
	return nil
}

// Upper adds upper bounds for parameters.
// These are enforced by a wrapper at the start of Run()
func (s *SplxCtrl) Upper(upper []float32) error {
	if err := checklen("upper", upper, s.iniPrm); err != nil {
		return err
	}
	s.upper = upper
	return nil
}

// Lower adds lower bounds for parameters, as for Upper
func (s *SplxCtrl) Lower(lower []float32) error {
	if err := checklen("lower", lower, s.iniPrm); err != nil {
		return err
	}
	s.lower = lower
	return nil
}

// splx is a dynamically allocated matrix from the matrix package.
type splx struct {
	*matrix.FMatrix2d
}

// inipntClassic does a classic initialisation of the simplex. The first point
// is at the starting place. Subsequent vertices at p_i = p_0 + lambda*e_i
// where e_i is a unit in that dimension.
func (s *SplxCtrl) iniPntClassic(splx splx) {
	for i := range splx.Mat {
		copy(splx.Mat[i], s.iniPrm)
	}
	for i := 1; i <= len(s.iniPrm); i++ {
		splx.Mat[i][i-1] += s.span[i-1]
	}
}

// inipntSpread initialises the vertices as a set of points evenly spread
// in each dimension
func (s *SplxCtrl) iniPntSpread(splx splx) {
	ndim := len(s.iniPrm)
	nvtx := ndim + 1
	for i, spread := range s.span {
		incrmt := spread / float32(ndim)
		start := s.iniPrm[i] - spread/2
		for j := 0; j < nvtx; j++ {
			splx.Mat[j][i] = start + float32(j)*incrmt
		}
	}
	if s.noPermute == false { // Allow us to turn off the permuting of values
		for ip := 0; ip < ndim; ip++ { // permute elements in each dimension
			for j, val := range rand.Perm(nvtx) {
				tmp := splx.Mat
				tmp[val][ip], tmp[j][ip] = tmp[j][ip], tmp[val][ip]
			}
		}
	}
}

// iniPoints allocates space for the actual simplex (a matrix) and
// fills it with values.
func (s *SplxCtrl) iniPoints() splx {
	ndim := len(s.iniPrm)
	nvtx := ndim + 1
	//  We might be given an amount of scatter, or ranges for each param
	if s.span == nil {
		s.span = make([]float32, len(s.iniPrm))
		for i := range s.iniPrm {
			s.span[i] = s.scatter * s.iniPrm[i]
		}
	}
	s.scatter = float32(math.NaN())
	splx := splx{matrix.NewFMatrix2d(nvtx, ndim)}

	if s.iniType == IniClassic {
		s.iniPntClassic(splx)
	} else {
		s.iniPntSpread(splx)
	}
	return splx
}

// sWk holds the scratch arrays for sums and ranks.
type sWk struct {
	cost    CostFun // copy of cost function
	testfun CostFun // Function for testing
	hstryF  *os.File
	testF   *os.File
	y       []float32 // y values at each simplex point
	cntrd   []float32 // centroid of all points, except worst
	ptrial  []float32 // trial point used in amotry
	rank    []int     // sorted list of ranks of vertices
	ncycle  int32     // how many cycles did we do
	yr      float32   //  y-value of reflected point
}

// accept replaces the worst point with the candidate and updates the y value
func accept(splx splx, sWk *sWk, newpoint []float32, newY float32) {
	ihi := sWk.rank[0]
	copy(splx.Mat[ihi], newpoint)
	sWk.y[ihi] = newY
}

// amotry does the reflection, contraction depending on fac.
// We return the new y value and fill out the ptrial slice in the sWk
// structure.
func amotry(fac float32, sWk *sWk, cndt []float32) (float32, error) {
	ndim := len(sWk.cntrd)
	var ytry float32
	var err error
	for i := 0; i < ndim; i++ {
		sWk.ptrial[i] = (1+fac)*sWk.cntrd[i] - fac*cndt[i]
	}

	if ytry, err = sWk.cost(sWk.ptrial); err != nil {
		ytry = float32(math.NaN()) // on error, also make the y-value invalid
	}
	return ytry, err
}

// centroid updates the simplex centroid. This is the middle of the vertices
// but excluding the worst (highest)
func (s *sWk) centroid(splx splx) {
	ndim := len(s.cntrd)
	nvtx := ndim + 1
	rank := s.rank
	for i := 0; i < ndim; i++ {
		var sum float32
		for j := 1; j < nvtx; j++ {
			sum += splx.Mat[rank[j]][i]
		}
		s.cntrd[i] = sum / float32(ndim)
	}
}

// contract brings all points halfway towards the lowest point
func contract(splx splx, sWk *sWk) {
	nvtx := len(splx.Mat)
	pntLow := splx.Mat[sWk.rank[nvtx-1]] // best point
	for i := 0; i < nvtx-1; i++ {
		ix := sWk.rank[i]
		for j, val := range pntLow {
			splx.Mat[ix][j] = (splx.Mat[ix][j] + val) / 2.0
		}
	}
}

// newSwk sets up the arrays for a simplex work (sWk) structure
func (sWk *sWk) init(ndim int, cost CostFun) {
	npnt := ndim + 1
	sWk.y = make([]float32, npnt)
	sWk.rank = make([]int, npnt)
	sWk.cntrd = make([]float32, ndim)
	sWk.ptrial = make([]float32, ndim)
	sWk.cost = cost
}

// setupFirstStep calculates values at the initial simplex vertices
// and puts values into the rank slice.
func (sWk *sWk) setupFirstStep(splx splx) error {
	for i, v := range splx.Mat {
		var err error
		if sWk.y[i], err = sWk.cost(v); err != nil {
			return fmt.Errorf("initialising simplex: %w", err)
		}
	}
	for i := 0; i < len(sWk.rank); i++ {
		sWk.rank[i] = i
	}
	return nil
}

// updateContract updates the function values at all vertices, except
// the best one, which was not changed during contraction
func (sWk *sWk) updateContract(splx splx) error {
	nvtx := len(sWk.y)
	for i := 0; i < (nvtx - 1); i++ {
		var err error
		ix := sWk.rank[i]
		if sWk.y[ix], err = sWk.cost(splx.Mat[ix]); err != nil {
			return fmt.Errorf("updating after contract: %w", err)
		}
	}
	return nil
}

// fabs is my punishment for insisting on 32 bit floats
func fabs(x float32) float64 {
	return math.Abs(float64(x))
}

// converged returns true if we have converged. Use the criterion from
// numerical recipes first to decide based on the cost function at the
// best and worst values.  Optionally keep contracting if the pTol
// slice has been set.
func (sWk *sWk) converged(splx splx, tol float32, pTol []float32) bool {
	const tiny = 1e-40
	yhi := sWk.y[sWk.rank[0]]
	n := len(sWk.y)
	ylo := sWk.y[sWk.rank[n-1]]
	rtol := (2 * fabs(yhi-ylo)) / (fabs(yhi) + fabs(ylo) + tiny)
	if rtol > float64(tol) {
		return false
	}
	if pTol != nil { // Optional
		hiPnt := splx.Mat[sWk.rank[0]]   // This loop ensures that
		loPnt := splx.Mat[sWk.rank[n-1]] // the range in each dimension
		for i := range hiPnt {           // is not too large.
			if math.Abs(float64(hiPnt[i]-loPnt[i])) > float64(pTol[i]) {
				return false
			}
		}
	}
	return true
}

// dbgdetail is for debugging. It is part of a closure, so we can get to
// the values in the simplex and working arrays.
func dbgdetail(sWk *sWk, splx splx, s string, fdbg *os.File) {
	if s == "close" {
		fdbg.Close()
		return
	}
	ihi := sWk.rank[0]
	ilo := sWk.rank[len(sWk.y)-1]
	for _, v := range sWk.y {
		if _, err := fmt.Fprintf(fdbg, "%.2f, ", v); err != nil {
			panic("Writing to debug file")
		}
	}
	fmt.Fprintf(fdbg, "%v, %v", ihi, ilo)
	for i := range splx.Mat {
		for _, v := range splx.Mat[i] {
			fmt.Fprintf(fdbg, ", %.2f", v)
		}
	}
	fmt.Fprintf(fdbg, ", %s\n", s)
}

// history writes a line saying the step number, the cost (y-value) and
// the current parameters.
func (sWk sWk) history(ncycle int32, splx splx) (err error) {
	n := len(sWk.y)
	ylo := sWk.y[sWk.rank[n-1]]
	loPnt := splx.Mat[sWk.rank[n-1]]
	var outfmt string = "%v %v "
	var ffmt string = "%7.2f"
	if sWk.hstryF != nil {
		_, err = fmt.Fprintf(sWk.hstryF, outfmt, ncycle, ylo)
		for _, v := range loPnt {
			fmt.Fprintf(sWk.hstryF, ffmt, v)
		}
		fmt.Fprintf(sWk.hstryF, "\n")
	}
	if sWk.testF != nil && err == nil {
		ytest, err := sWk.testfun(loPnt)
		if err != nil {
			return err
		}
		fmt.Fprintf(sWk.testF, outfmt, ncycle, ytest)
		for _, v := range loPnt {
			fmt.Fprintf(sWk.testF, ffmt, v)
		}
		fmt.Fprintf(sWk.testF, "\n")
	}
	return err
}

// onerun is the inner call to the simplex. It is called on each start
// and restart.
func (s *SplxCtrl) onerun(sWk *sWk) (uint8, error) {
	ndim := len(s.iniPrm)
	npnt := ndim + 1
	splx := s.iniPoints()

	if err := sWk.setupFirstStep(splx); err != nil {
		return Bust, fmt.Errorf("Setting up for first step: %w", err)
	}
	stopReason := MaxSteps
	sWk.ncycle = s.maxstep
	for n := int32(0); n < s.maxstep; n++ {
		var err error
		oldlo := sWk.rank[npnt-1] // old best point
		sort.Slice(sWk.rank, func(i, j int) bool {
			return sWk.y[sWk.rank[i]] > sWk.y[sWk.rank[j]]
		})
		ilo := sWk.rank[npnt-1] // best point
		i2nd := sWk.rank[1]     // second best
		ihi := sWk.rank[0]      // worst (hi) point
		if sWk.hstryF != nil || sWk.testF != nil {
			if oldlo != ilo {
				if err = sWk.history(n, splx); err != nil {
					return Bust, err
				}
			}
		}
		stopmaybe := func() {}
		if n > 90 {
			stopmaybe()
		}
		if sWk.converged(splx, s.tol, s.pTol) {
			sWk.ncycle = n
			stopReason = Converged
			break
		}
		sWk.centroid(splx)
		var yr, ye, yc float32
		cndt := splx.Mat[ihi]
		if yr, err = amotry(alpha, sWk, cndt); err != nil {
			return Bust, err
		}

		if yr < sWk.y[i2nd] { //  either accept or try extension
			accept(splx, sWk, sWk.ptrial, yr)
			if yr >= sWk.y[ilo] {
				continue
			}
			if ye, err = amotry(gamma, sWk, cndt); err != nil { // try extend
				return Bust, err
			}
			if ye < yr {
				accept(splx, sWk, sWk.ptrial, ye)
			}
			continue
		} // Now the cases involving contraction
		if yr < sWk.y[ihi] { // outside contraction
			yc, _ = amotry(beta, sWk, sWk.ptrial)
			if yc <= yr {
				accept(splx, sWk, sWk.ptrial, yc)
				continue
			}
		} else { // inside 1D contraction
			ycc, _ := amotry(beta, sWk, splx.Mat[ihi]) // Have I got sign right ?
			if ycc < sWk.y[ihi] {
				accept(splx, sWk, sWk.ptrial, ycc)
				continue
			}
		}

		contract(splx, sWk) // last option, general contraction
		sWk.updateContract(splx)
	}
	ilo := sWk.rank[npnt-1]
	copy(s.iniPrm, splx.Mat[ilo])
	return stopReason, nil
}

// setupBounds is a wrapper (closure) around the original cost function.
// If a point exceeds a bound, we return maxfloat32 which means the point
// will be rejected.
// We only need to return the function value at the worst point + some
// positive delta, but this would require passing extra information into
// the cost function.
func (s *SplxCtrl) setupBounds() {
	if s.lower == nil && s.upper == nil {
		return // if we do not have any bounds, we do not need the wrapper
	}
	origCost := s.cost
	cost := func(x []float32) (float32, error) {
		if s.upper != nil {
			for i, v := range s.upper {
				if x[i] > v {
					return math.MaxFloat32, nil
				}
			}
		}
		if s.lower != nil {
			for i, v := range s.lower {
				if x[i] < v {
					return math.MaxFloat32, nil
				}
			}
		}
		return origCost(x)
	}
	s.cost = cost
}

// Run invokes the simplex maxstart times. Returns a structure of type
// Result and an error.
func (s *SplxCtrl) Run(maxstart int) (Result, error) {
	var sWk sWk
	var err error
	errRes := Result{BestPrm: nil, Ncycle: 0, StopReason: Bust}
	ndim := len(s.iniPrm)
	s.setupBounds()
	sWk.init(ndim, s.cost) // move the copying of testfun here
	s.cost = nil           // pointer has now been handed to the sWk structure
	if s.hstryFname != "" {
		if sWk.hstryF, err = os.Create(s.hstryFname); err != nil {
			return errRes, err
		}
		defer sWk.hstryF.Close()
	}
	sWk.testfun = s.testfun
	s.testfun = nil
	if sWk.testfun != nil {
		if sWk.testF, err = os.Create(s.testFname); err != nil {
			return errRes, err
		}
		defer sWk.testF.Close()
	}
	var stopReason uint8
	for mr := 0; mr < maxstart; mr++ {
		if stopReason, err = s.onerun(&sWk); err != nil {
			return errRes, err
		}
		for i, v := range s.span {
			s.span[i] = v / 2. // halve the range before next restart
		}
	}
	result := Result{
		BestPrm:    s.iniPrm,
		Ncycle:     sWk.ncycle,
		StopReason: stopReason,
	}
	return result, nil
}
