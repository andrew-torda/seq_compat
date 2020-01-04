// Calculate some geometries, lengths and angles

package geom

import (
	"math"
	"github.com/andrew-torda/goutil/pdb/cmmn"
)

const (
	mindist    = 2.6
	mindist2   = mindist * mindist
	maxdist    = 4.1 // max dist for c_alpha to c_alpha
	maxdist2   = maxdist * maxdist
	Brokendist = -99
)

// xyzhelper makes the code below a bit more compact. Returns distance
// squared in one dimension or bigdist if it is bigger than our limit.
func xyzhelper(r1, r2 float32) (float32, error) {
	r := r1 - r2
	r = r * r
	if r >= maxdist2 {
		return r, Error("too big")
	}
	return r, nil
}

type Error string

func (e Error) Error() string { return string(e) }

// xyzdist gets the distance between two points, but if it is
// bigger than Brokendist or smaller than mindist, it returns
// an error.
func XyzDist(x1, x2 cmmn.Xyz) (float32, error) {
	var xd, yd, zd float32
	var err error
	if xd, err = xyzhelper(x1.X, x2.X); err != nil {
		return xd, err
	}
	if yd, err = xyzhelper(x1.Y, x2.Y); err != nil {
		return yd, err
	}
	if zd, err = xyzhelper(x1.Z, x2.Z); err != nil {
		return zd, err
	}

	r := xd + yd + zd
	if r <= mindist2 {
		return r, Error("too small")
	}

	return float32(math.Sqrt(float64(r))), nil
}

//xyzDiff gets the difference of two vectors
func xyzDiff(start, end cmmn.Xyz) (diff cmmn.Xyz) {
	diff.X = end.X - start.X
	diff.Y = end.Y - start.Y
	diff.Z = end.Z - start.Z
	return diff
}

// Xyzangle takes three points and returns the angle between them
func XyzAngle(a, b, c cmmn.Xyz) (float32, error) {
	x1 := xyzDiff(b, a)
	x2 := xyzDiff(b, c)
	dotp := x1.X*x2.X + x1.Y*x2.Y + x1.Z*x2.Z
	len1 := float64(x1.X*x1.X + x1.Y*x1.Y + x1.Z*x1.Z)
	len2 := float64(x2.X*x2.X + x2.Y*x2.Y + x2.Z*x2.Z)
	cosalpha := float64(dotp) / (math.Sqrt(len1) * math.Sqrt(len2))
	if cosalpha > 1 && cosalpha < 1.01 { // numerical noise
		return 0.0, nil
	}
	if cosalpha < -1 && cosalpha > -1.01 {
		return math.Pi, nil
	}
	if cosalpha < -1 || cosalpha > 1 {
		return float32(math.NaN()), Error("Broken angle")
	}
	return float32(math.Acos(cosalpha)), nil
}

// vecProd returns the vector product of two vectors
func vecProd(u, v cmmn.Xyz) (res cmmn.Xyz) {
	res.X = u.Y*v.Z - u.Z*v.Y
	res.Y = u.Z*v.X - u.X*v.Z
	res.Z = u.X*v.Y - u.Y*v.X
	return res
}

// xyz dotprod returns the dot / scalar product of two vectors
func sclrProd(u, v cmmn.Xyz) float32 { return (u.X*v.X + u.Y*v.Y + u.Z*v.Z) }

// xyzLen2 gives us the length squared
func xyzLen2(v cmmn.Xyz) float32 { return (v.X*v.X + v.Y*v.Y + v.Z*v.Z) }

// xyzLen returns the vector length
func xyzLen(v cmmn.Xyz) float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

// XyzDhdrl takes four points and returns the dihedral angle
func XyzDhdrl(ii, jj, kk, ll cmmn.Xyz) float32 {
	r_ij := xyzDiff(ii, jj)
	r_kj := xyzDiff(kk, jj)
	r_kl := xyzDiff(kk, ll)
	var r_im, r_ln cmmn.Xyz
	{
		tmp := sclrProd(r_ij, r_kj)
		tmp = tmp / xyzLen2(r_kj)
		tmp_vec := cmmn.Xyz{X: tmp * r_kj.X, Y: tmp * r_kj.Y, Z: tmp * r_kj.Z}
		r_im = xyzDiff(r_ij, tmp_vec)
	}
	{
		tmp := sclrProd(r_kl, r_kj)
		tmp = tmp / xyzLen2(r_kj)
		tmp_vec := cmmn.Xyz{X: tmp * r_kj.X, Y: tmp * r_kj.Y, Z: tmp * r_kj.Z}
		r_ln = xyzDiff(tmp_vec, r_kl)
	}
	var tau float32
	{
		t_cos := float64(sclrProd(r_im, r_ln) / (xyzLen(r_im) * xyzLen(r_ln)))
		if t_cos > 1 {   // Numerical errors can catch us. If so, no need
			return 0.0   // to call acos()
		}
		if t_cos < -1 {
			return math.Pi
		}
		tau = float32(math.Acos(t_cos))
	}

	if sclrProd(r_ij, vecProd(r_kj, r_kl)) >= 0 {
		return (tau)
	}
	return (-tau)
}
