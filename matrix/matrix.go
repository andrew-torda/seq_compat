// Package matrix 7 feb 2018
// Make a 2D array for floats, but we will use it later for other types
// There are a few ways into this.
// You can declare a FMatrix2d. This will have zero space allocated.
// Before you use it, call Resize. This will do the allocation.
// You can call NewFMatrix2d with the right size. This will just give you
// a matrix to use. It is best if you have a one-off use. You can
// resize it if you like.
// If you want to use matrices whose size changes on iterations of a loop,
// then declare the matrix at the start. On each iteration call Resize.
// The innards will grow as necessary, using the same backing store,
// but adapting the slices as necessary.
// In the silly case of zero rows or columns, the correct behaviour is
// unclear. We do not throw any errors or panic.
// If you have zero columns, then we allocate a slice of length zero,
// but n_row rows and all the slices point to a nil slice. It will explode
// if you put something there, but then that is what happens any time you go
// over array bounds. If you have zero rows, then no space and no pointers are
// allocated and the Size function will fail.

package matrix

import (
	"fmt"
)

// FMatrix2d is a two dimensional array of float32's
type FMatrix2d struct {
	Mat      [][]float32
	fullData []float32
}

// fixSlices sets the pointers in a matrix.
// It is in its own function so we can call it for new objects
// or when resizing an old one.
func (mat *FMatrix2d) fixSlices(n_r, n_c int) {
	tmp := mat.fullData
	mat.Mat = make([][]float32, n_r)
	for i := range mat.Mat {
		mat.Mat[i] = tmp[:n_c]
		tmp = tmp[n_c:]
	}
}

// FMatrix2dSize takes a matrix and desired size. If the size is too small,
// it reallocates the backing array.
// If the size is big enough, but the dimensions are not right, it
// runs over the pointers and sets them.
// It will not reduce the space needed by a matrix.
func (mat *FMatrix2d) Resize(n_r, n_c int) *FMatrix2d {
	nrow, ncol := mat.Size()
	if nrow == n_r && ncol == n_c { // If the old dimensions were OK
		return mat //                  were OK just return
	}

	if a, b := mat.Size(); n_r*n_c > a*b { // Is new size bigger than old ?
		mat.fullData = make([]float32, n_r*n_c)
	}
	mat.fixSlices(n_r, n_c)

	return mat
}

// NewFMatrix2d gives us a two dimensional matrix of m x n.
// We do not need the pointer to the backing store for simple
// use, but we keep it in case we want to resize the matrix
func NewFMatrix2d(n_r, n_c int) *FMatrix2d {
	r := new(FMatrix2d)
	r.fullData = make([]float32, n_r*n_c)
	r.fixSlices(n_r, n_c)
	return r
}

// Size acts on a FMatrix2d pointer and returns the number of rows and
// number of columns
func (mat *FMatrix2d) Size() (nrow, ncol int) {
	if nrow = len(mat.Mat); nrow == 0 {
		return 0, 0
	}
	ncol = len(mat.Mat[0])
	return
}

// String acts on a FMatrix2d pointer and returns a string with the
// Matrix printed out in a form that might be useful for debugging.
func (mat *FMatrix2d) String() (s string) {
	for _, row := range mat.Mat {
		for _, col := range row {
			x := col
			if x < -50 {
				x = -50
			}
			if x > 50 {
				x = 50
			}
			s += fmt.Sprintf("%5v", x)
		}
		s += "\n"
	}
	return s
}

// BackingDataString returns a string with the contents of the underlying array.
// It is only exported so the test file can get to it.
func (mat *FMatrix2d) BackingDataString() (s string) {
	s = fmt.Sprintln(mat.fullData)
	return
}

// ------------------------------------------------------------

// BMatrix2d is a two dimensional array of bytes
type BMatrix2d struct {
	Mat      [][]byte
	fullData []byte
}

// fix_slices sets the pointers in a matrix.
// It is in its own function so we can call it for new objects
// or when resizing an old one.
func (mat *BMatrix2d) fixSlices(n_r, n_c int) {
	tmp := mat.fullData
	mat.Mat = make([][]byte, n_r)
	for i := range mat.Mat {
		mat.Mat[i] = tmp[:n_c]
		tmp = tmp[n_c:]
	}
}

// NewBMatrix2d gives us a two dimensional matrix of m x n.
// We do not need the pointer to the backing store for simple
// use, but we keep it in case we want to resize the matrix
func NewBMatrix2d(n_r, n_c int) *BMatrix2d {
	r := new(BMatrix2d)
	r.fullData = make([]byte, n_r*n_c)
	r.fixSlices(n_r, n_c)
	return r
}

// Size acts on a FMatrix2d pointer and returns the number of rows and
// number of columns
func (mat *BMatrix2d) Size() (nrow, ncol int) {
	if nrow = len(mat.Mat); nrow == 0 {
		return 0, 0
	}
	ncol = len(mat.Mat[0])
	return
}

// BMatrix2dSize takes a matrix and desired size. If the size is too small,
// it reallocates the backing array.
// If the size is big enough, but the dimensions are not right, it
// runs over the pointers and sets them.
// It will not reduce the space needed by a matrix.
func (mat *BMatrix2d) Resize(n_r, n_c int) *BMatrix2d {
	nrow, ncol := mat.Size()
	if nrow == n_r && ncol == n_c { // If the old dimensions were OK
		return mat //                  were OK just return
	}

	if a, b := mat.Size(); n_r*n_c > a*b { // Is new size bigger than old ?
		mat.fullData = make([]byte, n_r*n_c)
	}
	mat.fixSlices(n_r, n_c)

	return mat
}

// String acts on a BMatrix2d pointer and returns a string with the
// Matrix printed out in a form that might be useful for debugging.
func (mat *BMatrix2d) String() (s string) {
	for _, row := range mat.Mat {
		for _, col := range row {
			s += fmt.Sprintf("%4.1d ", uint8(col))
		}
		s += "\n"
	}
	return s
}
