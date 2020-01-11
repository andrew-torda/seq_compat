// 27 Dec 2019
// Package seq provides a simplex (Nelder and Meade) optimizer.
// Using J.A. Nelder, R. Mead, Comp. J., 7, 308-313
// J.C. Lagarias, J.A. Reeds, M.H. Wright, and P.E. Wright
// Press, W.H., Teukolsky, S.A., Vetterling, W.T., Flannery, B.P.,
// Numerical Recipes in C., Cambridge University Press, 1992
// The structure with the amotry() function comes from numerical recipes,
// but the formulae for moving the highest point around are taken from
// the primary references.
// It has a couple of frills.
//  1. It scrambles the order of coordinates in the original simplex.
//     This has minimises the effect of passenger (unimportant) coordinates.
//  2. It allows a vector of minimum and maximum values. It will reject
//     moves if they go beyond these boundaries. This is done by wrapping
//     the cost function.
// Could be improved
// * we call a full sort after each cycle. This is not necessary. One
//   only needs a list with the highest, next-highest and best points.
// * We re-calculate the centroid on each cycle. This is necessary, but
//   the version in numerical recipes does it by removing the old best
//   value and adding in the new one. In a few dimensions, it makes no
//   difference. In many dimensions, one could argue the method in numerical
//   recipes will be faster.
// What I am working on, to do
//   * at the start, check the initial point does not exceed any bounds.
//     Afterwards, this means that we can set y=y[ihi]+1 if we exceed a
//     bound.
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
type CostFun func(x []float32) (float32, error)
type SplxCtrl struct {
	maxstep   int       // cycles of the simplex
	lower     []float32 // bounds on the parameters
	upper     []float32
	iniPrm    []float32 // initial parameter guess
	span      []float32 // initial range for each parameter
	pTol      []float32 // optional tolerance for each parameter.
	seed      int       // Seed for random number generator
	scatter   float32   // Spread around initial simplex points
	cost      CostFun   // the function to be optimised
	testfun   CostFun
	tol       float32
	ncycle    int  // how many cycles did we do
	noPermute bool // turn off permuting of simplex at setup
}

// Result is the structure used to pass results back to the caller
type Result struct {
	Ncycle     int       // How many times did we cycle
	BestPrm    []float32 // best parameters found
	StopReason uint8     // Reason why calculation stopped
}

// NewSimplex just gives us a structure with default values.
// The cost function must be specified.
func NewSplxCtrl(cost CostFun, iniPrm []float32, maxstep int) *SplxCtrl {
	s := new(SplxCtrl)
	s.cost = cost
	s.iniPrm = iniPrm
	s.maxstep = maxstep
	s.tol = 1e-8
	s.seed = randSeed
	s.scatter = 0.1 // 10 % scatter means original +/- 5 %
	rand.Seed(int64(s.seed))
	return s
}

// checklen is used in the set functions which take a slice (usually of float32's).
func checklen (name string, got []float32, want []float32) error {
	ngot := len (got)
	nwant := len (want)
	if ngot != nwant {
		e := fmt.Errorf ("checking parameter \"name\". Length mismatch. Got %d. Wanted %d",
			ngot, nwant)
		return e
	}
	return nil
}

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
	if err := checklen ("span", x, s.iniPrm); err != nil {
		return err
	}
	s.span = x
	return nil
}

// Ptol sets a tolerance for each parameter. It is optional. If the tol criterion is met,
// then we can ask in each dimension if the x_min - x_max is smaller than some value.
// This is only useful for noisy cost functions and will make the simplex keep
// contracting a bit more.
func (s *SplxCtrl ) Ptol (pTol []float32) error {
	if err := checklen ("pTol", pTol, s.iniPrm); err != nil {
		return err
	}
	s.pTol = pTol
	return nil
}
	

// Upper adds upper bounds for parameters.
// These are enforced by a wrapper at the start of Run()
func (s *SplxCtrl) Upper(upper []float32) error {
	if err := checklen ("upper", upper, s.iniPrm); err != nil {
		return err
	}	
	s.upper = upper
	return nil
}

// Lower adds lower bounds for parameters, as for Upper
func (s *SplxCtrl) Lower(lower []float32) error {
	if err := checklen ("lower", lower, s.iniPrm); err != nil { return err}
	s.lower = lower
	return nil
}

// splx is a dynamically allocated matrix from the matrix package.
type splx struct {
	*matrix.FMatrix2d
}

// iniPoints allocates space for the actual simplex (a matrix) and
// fills it with values.
func (s *SplxCtrl) iniPoints() splx { //*matrix.FMatrix2d {
	nparam := len(s.iniPrm)
	npoint := nparam + 1
	//  We might be given an amount of scatter, or ranges for each param
	if s.span == nil {
		s.span = make([]float32, len(s.iniPrm))
		for i := range s.iniPrm {
			s.span[i] = s.scatter * s.iniPrm[i]
		}
		s.scatter = float32(math.NaN())
	}

	splx := splx{matrix.NewFMatrix2d(npoint, nparam)}

	for i, spread := range s.span {
		incrmt := spread / float32(nparam)
		start := s.iniPrm[i] - spread/2
		for j := 0; j < npoint; j++ {
			splx.Mat[j][i] = start + float32(j)*incrmt
		}
	}
	if s.noPermute { // for debugging, we may want to use
		return splx // simplex unpermuted
	}

	for ip := 0; ip < nparam; ip++ { // permute elements in each dimension
		for j, val := range rand.Perm(npoint) {
			tmp := splx.Mat
			tmp[val][ip], tmp[j][ip] = tmp[j][ip], tmp[val][ip]
		}
	}
	return splx
}

// sWk holds the scratch arrays for sums and ranks.
type sWk struct {
	cost      CostFun   // copy of cost function
	y         []float32 // y values at each simplex point
	cntrd     []float32 // centroid of all points, except worst
	ptrial    []float32 // trial point used in amotry
	rank      []int     // sorted list of ranks of vertices
	dbgDetail func(string)
}

const (
	yesImprove uint8 = iota // reflection or expansion improved worst point
	noImprove
	bustImprove // on error, something bust
)

// amotry moves the worst vertex by reflection, expansion or 1D contraction
// as determined by fac.
func amotry(splx splx, fac float32, sWk *sWk) (uint8, error) {
	ndim := len(sWk.cntrd)
	ihi := sWk.rank[0]
	i2nd := sWk.rank[1] // second best point

	for i := 0; i < ndim; i++ {
		sWk.ptrial[i] = (1+fac)*sWk.cntrd[i] - fac*splx.Mat[ihi][i]
	}
	var ytry float32
	var err error
	if ytry, err = sWk.cost(sWk.ptrial); err != nil {
		return bustImprove, err
	}
	if ytry >= sWk.y[i2nd] {
		return noImprove, nil
	}
	copy(splx.Mat[ihi], sWk.ptrial) // save the yval and the trial move
	sWk.y[ihi] = ytry

	return yesImprove, nil
}

// centroid updates the simplex centroid. This is the middle of the points,
// but excluding the worst (highest)
func (s *sWk) centroid(splx splx) {
	ndim := len(s.cntrd)
	npoint := ndim + 1
	rank := s.rank
	for i := 0; i < ndim; i++ {
		var sum float32
		for j := 1; j < npoint; j++ {
			sum += splx.Mat[rank[j]][i]
		}
		s.cntrd[i] = sum / float32(ndim)
	}
}

// contract brings all points halfway towards the lowest point
func contract(splx splx, sWk *sWk) {
	npoint := len(splx.Mat)
	pntLow := splx.Mat[sWk.rank[npoint-1]] // best point
	for i := 0; i < npoint-1; i++ {
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

// setupfirststep calculates values at the initial simplex vertices,
// sorts them and so on.
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
	npoint := len(sWk.y)
	for i := 0; i < (npoint - 1); i++ {
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

// converged returns true if we have converged. Use the criterion from numerical
// recipes first to decide based on the cost function at the best and worst values.
// Optionally keep contracting if the pTol slice has been set.
func (sWk *sWk) converged(splx splx, tol float32, pTol []float32) bool {
	const tiny = 1e-40
	yhi := sWk.y[sWk.rank[0]]
	n := len(sWk.y)
	ylo := sWk.y[sWk.rank[n-1]]
	rtol := (2 * fabs(yhi-ylo)) / (fabs(yhi) + fabs(ylo) + tiny)
	if rtol > float64(tol) {
		return false
	}
	if pTol != nil {
		hiPnt := splx.Mat[sWk.rank[0]]
		loPnt := splx.Mat[sWk.rank[n - 1]]
		for i := range hiPnt {
			if math.Abs(float64(hiPnt[i] - loPnt[i])) > float64(pTol[i]) {
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

// onerun is the inner call to the simplex. It will be called with
// different starting points on each call.
func (s *SplxCtrl) onerun(sWk *sWk) (uint8, error) {
	ndim := len(s.iniPrm)
	npnt := ndim + 1
	var fdbg *os.File
	{
		var err error
		if fdbg, err = os.Create("dbg"); err != nil {
			panic("debugging code, create failure: " + err.Error())
		}
	}
	splx := s.iniPoints()
	sWk.dbgDetail = func(s string) {
		dbgdetail(sWk, splx, s, fdbg)
	}

	if err := sWk.setupFirstStep(splx); err != nil {
		return Bust, fmt.Errorf("Setting up for first step: %w", err)
	}
	stopReason := MaxSteps
	s.ncycle = s.maxstep
	for n := 0; n < s.maxstep; n++ {
		var tRes uint8
		var err error
		sort.Slice(sWk.rank, func(i, j int) bool {
			return sWk.y[sWk.rank[i]] > sWk.y[sWk.rank[j]]
		})
		ilo := sWk.rank[npnt-1] // best point
		ihi := sWk.rank[0]      // worst (hi) point
		if sWk.converged(splx, s.tol, s.pTol) {
			s.ncycle = n
			sWk.dbgDetail("close")
			stopReason = Converged
			break
		}
		sWk.centroid(splx)
		sWk.dbgDetail("n")
		if tRes, err = amotry(splx, alpha, sWk); err != nil {
			return Bust, err
		}
		if tRes == yesImprove {
			if sWk.y[ihi] > sWk.y[ilo] {
				sWk.dbgDetail("r")
				continue // just accept and move on
			} // next, try extend
			if tRes, err = amotry(splx, gamma, sWk); err != nil {
				return Bust, err
			}
			if tRes == yesImprove { // expansion worked
				sWk.dbgDetail("e")
				continue
			}
		} else { // 1D contract and then general contract
			if tRes, err = amotry(splx, beta, sWk); err != nil {
				return Bust, err
			}
			if tRes == yesImprove {
				sWk.dbgDetail("c")
				continue // 1 point contraction worked
			}
			contract(splx, sWk) // last option, general contraction
			sWk.updateContract(splx)
			sWk.dbgDetail("a")
		}

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
		return
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

// Run will do a run
func (s *SplxCtrl) Run(maxstart int) (Result, error) {
	var sWk sWk
	ndim := len(s.iniPrm)
	s.setupBounds()
	sWk.init(ndim, s.cost)
	s.cost = nil // pointer has now been handed to the sWk structure
	var stopReason uint8
	for mr := 0; mr < maxstart; mr++ {
		var err error
		if stopReason, err = s.onerun(&sWk); err != nil {
			return Result{BestPrm: nil, Ncycle: 0, StopReason: Bust}, err
		}
		s.scatter /= 4.
	}
	result := Result{
		BestPrm:    s.iniPrm,
		Ncycle:     s.ncycle,
		StopReason: stopReason,
	}
	return result, nil
}
