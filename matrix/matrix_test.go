package matrix_test

import (
	"fmt"
	. "github.com/andrew-torda/goutil/matrix"
	"os"
	"testing"
)

// Just for testing
const (
	nrowDflt = 3
	ncolDflt = 5
)

var testSizes = []struct {
	nr, nc int
}{
	{5, 0},
	{0, 0},
	{3, 5},
	{5, 3},
	{5, 3},
	{4, 4},
	{1, 1},
}

func fill_access(mat *FMatrix2d, nr, nc int) {
	var n float32 = 1
	for i := 0; i < nr; i++ {
		for j := 0; j < nc; j++ {
			mat.Mat[i][j] = n
			n++
		}
	}
}

// Check a matrix if it seems to be the right size
func checkMat(m *FMatrix2d, nr int, nc int, t *testing.T) {
	if nrow, ncol := m.Size(); nrow != nr || ncol != nc {
		t.Fatal("TestSize rows x cols, wanted", nr, nc, "got", nrow, ncol)
	}
	fill_access(m, nr, nc)
}

// Make a fresh matrix on each invocation
func TestFresh(t *testing.T) {
	for _, sizes := range testSizes {
		fmt.Fprintln(os.Stderr, "sizes are ", sizes)
		m := NewFMatrix2d(sizes.nr, sizes.nc)
		checkMat(m, sizes.nr, sizes.nc, t)
	}
}

// TestNoInit call the matrix resize on a matrix not initialised
func TestNoInit(t *testing.T) {
	for _, sizes := range testSizes {
		var m FMatrix2d
		m.Resize(sizes.nr, sizes.nc)
		checkMat(&m, sizes.nr, sizes.nc, t)
	}
}

// TestResize make a matrix and resize it a few times
func TestResize(t *testing.T) {
	m := NewFMatrix2d(0, 0)
	for _, sizes := range testSizes {
		m.Resize(sizes.nr, sizes.nc)
		checkMat(m, sizes.nr, sizes.nc, t)
	}
}
