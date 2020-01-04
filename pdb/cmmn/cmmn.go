// Package pdb/cmmn has common definitions for coordinates and
// pdb files
package cmmn

import (
	"math"
)

// Does our data come from a file or http source ?
const (
	FileSrc byte = iota
	HTTPSrc
)

type Xyz struct{ X, Y, Z float32 }
type XyzSl []Xyz // xyz's are coordinates
type CoordSet map[string]XyzSl

var BrokenXyz = Xyz{math.MaxFloat32, 0, -math.MaxFloat32}

var BrokenResNum int = -9999

func (xyz *Xyz) Ok() bool {
	if *xyz != BrokenXyz {
		return true
	}
	return false
}

// A simple structure for one model, one chain and a set of xyz coordinates
type Chain struct {
	ChainID  string // Name, like "A" or "B"
	MdlNum   int16  // Model number
	NumLbl   []int  // residue numbers from file. Not real indices
	InsCode  []byte // Insertion code
	CoordSet CoordSet
}

// This is obviously just a slice of chains, but we have to define a type
// if we want to define a method on it
type ChnSl []Chain
// ChainNames returns a slice with the names of the chains.
func (chns ChnSl) ChainNames () (ret []string) {
	ret = make([]string, len(chns), 1)
	for i, k := range chns {
		ret[i] = k.ChainID
	}
	return
}
