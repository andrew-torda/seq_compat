// I need to do comparisons of atom names, but these are usually strings
// of length 1, 2 or sometimes 3. Check if it is faster to
// a. use a map / hash
// b. use simple string comparison
// c. special case comparisons of 1, 2, 3 or more.
// The winner is always the special case.
// For a small set of interesting strings (about 5) second place is a
// direct a==b string compare. For a set of six, map just gets second place.

package mmcif

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
)

const rndSize = 2000000

var rndName [rndSize]string
var intrst = []string{"CA", "C", "CB", "N", "O", "XXx"}
var decoys = []string{"CG", "CD", "CE1", "CZ1"}
var tstatom []string
var map1 map[string]bool

func init() {
	r := rand.New(rand.NewSource(1637))
	map1 = make(map[string]bool)
	tstatom = intrst
	tstatom = append(tstatom, decoys...)
	for i := range rndName {
		v := r.Intn(len(tstatom))
		rndName[i] = tstatom[v]
	}
	for _, s := range intrst {
		map1[s] = true
	}
}

func boringAtomDrct(atName string, names []string) bool {
	for _, s := range names {
		if atName == s {
			return false
		}
	}
	return true
}

func boringAtomSwtch(atName string, names []string) bool {
	for _, s := range names {
		if len(atName) != len(s) {
			continue
		}
		switch len(s) {
		case 1:
			if atName[0] == s[0] {
				return false
			}
		case 2:
			if atName[0] == s[0] && atName[1] == s[1] {
				return false
			}
		case 3:
			if atName[0] == s[0] && atName[1] == s[1] && atName[2] == s[2] {
				return false
			}
		default:
			if atName == s {
				return false
			}
		}
	}
 	return true
}

func boringAtomMap(atName string, names []string) bool {
	if _, ok := map1[atName]; ok == false {
		return true
	}
	return false
}

const nRep int = 50

func BenchmarkCmpDrct(b *testing.B) {
	var t bool
	for j := 0; j < nRep; j++ {
		for _, s := range rndName {
			t = boringAtomDrct(s, intrst)
		}
	}
	fmt.Fprintln(ioutil.Discard, t)
}

func BenchmarkCmpSwtch(b *testing.B) {
	var t bool
	for j := 0; j < nRep; j++ {
		for _, s := range rndName {
			t = boringAtomSwtch(s, intrst)
		}
	} 
	fmt.Fprintln(ioutil.Discard, t)
}

func BenchmarkCmpMap(b *testing.B) {
	var t bool
	for j := 0; j < nRep; j++ {
		for _, s := range rndName {
			t = boringAtomMap(s, intrst)
		}
	}
	fmt.Fprintln(ioutil.Discard, t)
}
