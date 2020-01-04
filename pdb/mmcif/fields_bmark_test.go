package mmcif

import (
	"bytes"
	"testing"
)

var s = []string{
	"ATOM   2    C  CA  A MET A 1 1   ? 24.415 9.736   -9.941  0.77 40.13 ? 0   MET A CA  1",
	"ATOM   3    C  C   A MET A 1 1   ? 25.817 10.178  -10.333 0.77 38.80 ? 0   MET A C   1 ",
	"ATOM   4    O  O   A MET A 1 1   ? 26.797 9.485   -10.058 0.77 39.07 ? 0   MET A O   1 ",
	"ATOM   5    C  CB  A MET A 1 1   ? 24.003 10.413  -8.634  0.77 42.21 ? 0   MET A CB  1 ",
}

var res []BSlice
var r2  [][]byte

func BenchmarkLibFields(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, ss := range s {
			r2 = bytes.Fields([]byte(ss))
		}
	}
}

func BenchmarkFields1(b *testing.B) {
	var scrtch [40]BSlice
	for i := 0; i < b.N; i++ {
 		for _, ss := range s {
			res = fields([]byte(ss), scrtch[:])
		}
	}
}
func BenchmarkFields3(b *testing.B) {
	var scrtch [40]BSlice
	for i := 0; i < b.N; i++ {
		for _, ss := range s {
			res = fields([]byte(ss), scrtch[:])
		}
	}
}
