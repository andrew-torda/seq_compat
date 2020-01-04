// 27 Dec 2019
// Package seq provides a simplex (Nelder and Meade) optimizer.
// It has a couple of frills.
//  1. It scrambles the order of coordinates in the original simplex.
//     This has minimises the effect of passenger (unimportant) coordinates.
//  2. It expects a vector of minimum and maximum values. It will reject
//     moves if they go beyond these boundaries

// What I am working on, to do
// * wrapper for cost function. If we have no upper and lower bounds,
//   cost is just cost
// * if we have bounds, then we use a wrapper. If no points are bounded,
//   then call cost. If we have bounds, then check, if a point is out of
//   bounds, then just return +inf or y[rank[0]] + 1.0

package simplex

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"github.com/andrew-torda/goutil/matrix"
)

const (
	Converged = iota
	Maxsteps
	Bounderror
)

const (
	randSeed         = 1637 // default seed for permuting
	alpha    float32 = 1
	beta     float32 = -1 / 2.
	gamma    float32 = -2
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
	maxstep  int       // cycles of the simplex
	maxstart int       // complete restarts
	lower    []float32 // bounds on the parameters
	upper    []float32
	bestPrm  []float32 // best params found so far
	seed     int       // Seed for random number generator
	scatter  float32   // Spread around initial simplex points
	cost     CostFun   // the function to be optimised
	testfun  CostFun
	noPermute bool     // turn off permuting of simplex at setup
}

// NewSimplex just gives us a structure with default values.
// The cost function must be specified.
func NewSplxCtrl(cost CostFun, iniPrm []float32) *SplxCtrl {
	s := new(SplxCtrl)
	s.cost = cost
	s.bestPrm = iniPrm
	s.seed = randSeed
	s.scatter = 0.1 // 10 % scatter means original +/- 5 %
	rand.Seed(int64(s.seed))
	return s
}

func (s *SplxCtrl) nopermute () {
	s.noPermute = true
}

// AddBounds lets one add a slice of lower or upper bounds for parameters.
// If you specify bounds for one parameter, you have to specify bounds for
// all.
func (s *SplxCtrl) AddBounds(lower []float32, upper []float32) error {
	if len(lower) != 0 {
		if len(lower) != len(s.bestPrm) {
			return errors.New("lower bounds wrong dimensions")
		}
		s.lower = lower
	}
	if len(upper) != 0 {
		if len(upper) != len(s.bestPrm) {
			return errors.New("upper bounds wrong dimensions")
		}
		s.upper = upper
	}
	return nil
}

// splx is really just a dynamically allocated matrix from the matrix package.
type splx struct {
	*matrix.FMatrix2d
}

// iniPoints will make our simplex and fill it with initial points
func (s *SplxCtrl) iniPoints() splx { //*matrix.FMatrix2d {
	nparam := len(s.bestPrm)
	npoint := nparam + 1

	splx := splx {matrix.NewFMatrix2d(npoint, nparam)}
	for i := 0; i < nparam; i++ {
		spread := s.bestPrm[i] * s.scatter
		incrmt := spread / float32(nparam)
		start := s.bestPrm[i] - spread/2
		for j := 0; j < npoint; j++ {
			splx.Mat[j][i] = start + float32(j)*incrmt
		}
	}
	if s.noPermute {  // for debugging, we may want to use
		return splx // simplex unpermuted
	}
	for ip := 0; ip < nparam; ip++ { // permute elements in each dimension
		for j, val := range rand.Perm(npoint) {
			splx.Mat[val][ip], splx.Mat[j][ip] = splx.Mat[j][ip], splx.Mat[val][ip]
		}
	}
	return splx
}

// sWk holds the scratch arrays for sums and ranks.
type sWk struct {
	cost   CostFun   // copy of cost function
	y      []float32 // y values at each simplex point
	cntrd  []float32 // centroid of all points, except worst
	ptrial []float32 // trial point used in amotry
	rank   []int     // sorted list of ranks of vertices
}

func nothing(a interface{}) {}

const (
	noImprove uint8 = iota
	yesImprove
	bustImprove
)

// amotry tries to reflect the worst point through the face of the simplex.
// If it succeeds, it tries to extend the operation.
// This operation is the same, whether we are reflecting, extending or
// doing a local contraction.
func amotry(splx splx, fac float32, sWk sWk) (uint8, error) {
	ndim := len(sWk.cntrd)
	ihi := sWk.rank[0]
	for i := 0; i < ndim; i++ {
		sWk.ptrial[i] = (1+fac)*sWk.cntrd[i] - fac*splx.Mat[ihi][i]
	}
	if ytry, err := sWk.cost(sWk.ptrial); err != nil {
		return bustImprove, err
	} else { // save the yval and the trial move
		if ytry > sWk.y[ihi] { // no improvement
			return noImprove, nil
		}
		copy(splx.Mat[ihi], sWk.ptrial)
		sWk.y[ihi] = ytry
	}
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
func contract(splx splx, sWk sWk) {
	npoint := len(splx.Mat)
	pntLow := splx.Mat[sWk.rank[npoint-1]] // best point
	for i := 0; i < npoint-1; i++ {
		ix := sWk.rank[i]
		for j, val := range pntLow {
			splx.Mat[ix][j] = (splx.Mat[ix][j] + val) / 2.0
		}
	}
}

// onerun is the inner call to the simplex. It will be called with
// different starting points on each call.
func (s *SplxCtrl) onerun(maxstep int) error {
	var sWk sWk

	splx := s.iniPoints()
	ndim := len(s.bestPrm)
	npnt := ndim + 1
	sWk.y = make([]float32, npnt)
	sWk.rank = make([]int, npnt)
	sWk.cntrd = make([]float32, ndim)
	sWk.ptrial = make([]float32, ndim)
	sWk.cost = s.cost
	for i, v := range splx.Mat {
		var err error
		if sWk.y[i], err = s.cost(v); err != nil {
			return fmt.Errorf("initialising simplex: %w", err)
		}
	}
	for i := 0; i < len(sWk.rank); i++ {
		sWk.rank[i] = i
	}

	for n := 0; n < s.maxstep; n++ {
		var tRes uint8
		var err error
		ilo := sWk.rank[npnt-1] // best point
		ihi := sWk.rank[0]      // worst (hi) point
		//		inxt := sWk.rank[1]     // next-best point
		sort.Slice(sWk.rank, func(i, j int) bool {
			return sWk.y[sWk.rank[i]] < sWk.y[sWk.rank[j]]
		})
		sWk.centroid(splx)
		if tRes, err = amotry(splx, alpha, sWk); err != nil {
			return err
		}
		if tRes == yesImprove {
			if sWk.y[ihi] > sWk.y[ilo] {
				continue // just accept and move on
			} // next, try extend
			if tRes, err = amotry(splx, gamma, sWk); err != nil {
				return err
			}
			if tRes == yesImprove { // expansion worked
				continue
			}
		} else { // 1D contract and then general contract
			if tRes, err = amotry(splx, beta, sWk); err != nil {
				return err
			}
			if tRes == yesImprove {
				continue // 1 point contraction worked
			}
			contract(splx, sWk) // last option, general contraction
		}

	}
	return nil
}

// Run will do a run
func (s *SplxCtrl) Run(maxstep, maxstart int) error {
	s.maxstep = maxstep
	s.maxstart = maxstart
	for mr := 0; mr < maxstart; mr++ {
		s.onerun(maxstep) // and here we give it new starting point
		s.scatter /= 2.
	}
	return nil
}
