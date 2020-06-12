// calpha will read pdb files and get some statistics about alpha carbons.
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/andrew-torda/goutil/pdb"
	"github.com/andrew-torda/goutil/pdb/calpha/geom"
	"github.com/andrew-torda/goutil/pdb/cmmn"
)

const (
	nReaderDflt = 3
	maxDirDflt  = 0
	exitSuccess = 0
	exitFail    = 1

	conv       = 180 / math.Pi
	parentPath = "/work/public/no_backup/pdb/data/structures/divided/mmCIF/"
)

type fresult struct {
	nbyte int64
	nfile int
}

func breaker() {}

// readpdb reads a file fullpath and returns a slice with alpha carbon
// coordinates - an XyzSl
func readpdb(fullpath string) ([]cmmn.XyzSl, error) {
	fp, err := os.Open(fullpath)
	if err != nil {
		return nil, err
	}
	var ret []cmmn.XyzSl
	defer fp.Close()
	if p, err := pdb.ReadCoord(fullpath, cmmn.FileSrc, "testoutput"); err != nil {
		return nil, err
	} else {
		breaker()
		for _, v := range p {
			c := v.CoordSet
			ca := c["CA"]
			var n_ca uint8 = 0
			for _, xyz := range ca {
				if xyz != cmmn.BrokenXyz {
					n_ca++        // Check if we have at least
					if n_ca > 1 { // two alpha carbons in this chain
						break
					}
				}
			}
			if n_ca > 1 {
				ret = append(ret, ca)
			}
		}
	}
	return ret, nil
}

// This is for debugging. Delete this declaration and let the compiler
// find the places it is no longer needed.
var namechan chan string

// getstat reads slices of coordinates from a channel.
func getstat(cXyz <-chan cmmn.XyzSl, wgGetstat *sync.WaitGroup) {
	defer wgGetstat.Done()
	for chain := range cXyz {
		_ = <-namechan

		var ii, jj, kk int
		for ii = 0; ii < len(chain)-2; ii++ { // skip leading broken
			if chain[ii] != cmmn.BrokenXyz {
				break
			}
		}
		for jj = ii + 1; jj < len(chain)-1; jj++ {
			if chain[jj] != cmmn.BrokenXyz {
				break
			}
		}
		for kk = jj + 1; kk < len(chain); kk++ {
			if chain[kk] != cmmn.BrokenXyz {
				break
			}
		}
		var r1 float32
		var err error
		var skip_next bool
		for i, j, k := ii, jj, kk; k < len(chain); {
			if skip_next {
				skip_next = false
				continue
			}
			var angle float32
			xI, xJ, xK := chain[i], chain[j], chain[k]
			if r1, err = geom.XyzDist(xI, xJ); err == nil {
				if _, err = geom.XyzDist(xJ, xK); err == nil {

					if angle, err = geom.XyzAngle(xI, xJ, xK); err == nil {
						fmt.Printf("%.2f,%.2f\n", r1, angle*conv)
					} else { // second distance is broken
						skip_next = true
					}
				}
			} else { // r1 was not ok
				// do nothing, just advance
			}
			i, j = j, k
			for k = k + 1; k < len(chain); k++ {
				if chain[k] == cmmn.BrokenXyz {
					continue
				}
				break
			}
		}
	}

}

// readPdir takes the name of a directory (PDB) from a channel.
// It opens the directory and then calls readpdb on every file
// it finds.
func readPdir(ch <-chan string, cXyz chan cmmn.XyzSl, res chan<- fresult,
	wgpdir *sync.WaitGroup, parentpath string) {
	defer wgpdir.Done()
	var fres fresult

	for d := range ch { // d is a directory name
		dd := parentpath + d
		fp, err := os.Open(parentpath + d)
		if err != nil {
			fmt.Println("Ignoring", dd, err)
			continue
		}
		files, err := fp.Readdirnames(0)
		fp.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, dd, err)
		}
		for _, file := range files {
			fullpath := dd + "/" + file
			caSl, err := readpdb(fullpath)
			if err != nil {
				fmt.Fprintln(os.Stderr, file, err)
			}
			for _, chain := range caSl {
				cXyz <- chain
				namechan <- filepath.Base(fullpath)
			}
		}
	}
	res <- fres
}

func mymain() (retval int) {
	var nReader int
	var maxDir int
	flag.IntVar(&nReader, "r", nReaderDflt, "num reader threads")
	flag.IntVar(&maxDir, "d", maxDirDflt, "max num directories to read")
	flag.Parse()
	fp, err := os.Open(parentPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitFail
	}
	defer fp.Close()

	dirs, err := fp.Readdirnames(0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error ", err)
		return exitFail
	}

	cdir := make(chan string, 200) // channel for directory names
	res := make(chan fresult)      // channel to send results back
	cXyz := make(chan cmmn.XyzSl)  // used to send xyz coord slices
	namechan = make(chan string)
	go func() {
		n := 0
		for _, file := range dirs {
			if maxDir == 0 || n < maxDir {
				cdir <- file
			}
			n++
		}
		close(cdir)
	}()

	var wgGetstat sync.WaitGroup
	go getstat(cXyz, &wgGetstat)
	wgGetstat.Add(1)

	var wgpdir sync.WaitGroup
	for i := 0; i < nReader; i++ {
		wgpdir.Add(1)
		go readPdir(cdir, cXyz, res, &wgpdir, parentPath)
	}

	var totalres fresult
	for i := 0; i < nReader; i++ {
		f := <-res
		//		fmt.Println("nbyte", f.nbyte, "nfile", f.nfile)
		totalres.nbyte += f.nbyte
		totalres.nfile += f.nfile
	}
	const mb = 1024 * 1024
	//	nbyte := float32(totalres.nbyte) / float32(mb)
	//	fmt.Printf("Totals nbyte %.2f Mb nfiles %d\n", nbyte, totalres.nfile)
	wgpdir.Wait()
	close(cXyz)
	close(namechan)
	wgGetstat.Wait()

	return exitSuccess
}

func main() {
	os.Exit(mymain())
}
