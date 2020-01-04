package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"../cmmn"
	"pdb/mmcif"
	"pdb/zwrap"
	"sort"
	"sync"
)

// Read a filename from a channel, read from it and send results on the
// results chan.

type rslt struct {
	name string
	err  error
}

func eatPDB(fname string, nummap nummap) (err error) {
	var rdr io.ReadCloser

	if r1, err := os.Open(fname); err != nil {
		return errorName(fname, err)
	} else {
		if rdr, err = zwrap.WrapMaybe(r1); err != nil {
			return errorName(fname, err)
		}
	}
	defer rdr.Close()
	mr := mmcif.NewMmcifReader(rdr)
	mr.SetAtoms(nil)
	mr.SetChains(nil)
	md, err := mr.DoFile()
	if err != nil {
		return errorName(fname, err)
	}

	cs := md.GetChains()
	for _, c := range cs {
		cset := c.CoordSet
		for k, v := range cset {
			for _, vv := range v {
				if vv != cmmn.BrokenXyz { // is atom valid ?
					nummap[k] = nummap[k] + 1
				}
			}
		}
	}
	return nil
}

// pdbStat operates on just one filename that it gets from channel nmChan
// (names channel).
// Sends results back on rslt
func pdbStat(nmChan chan fileRes, nummap nummap, wg *sync.WaitGroup, errflag *int) {
	defer wg.Done()
	for f := range nmChan {
		if f.err != nil { // A single error on the channel with names is
			fmt.Fprintln(os.Stderr, "name channel error ", f.name, f.err.Error())
			(*errflag)++
			return
		}

		if err := eatPDB(f.name, nummap); err != nil {
			(*errflag)++
			fmt.Fprintln(os.Stderr, "Error on ", f.name, err.Error())
			if *errflag >= 10 {
				return
			}
		}
	}

}

type nummap map[string]int

// collectData reads from the channel that the writers are writing to.
func collectData(nmChan chan fileRes, opts opts) (map[string]int, error) {
	var wg sync.WaitGroup
	var errflag int
	cmaps := make([]nummap, opts.nReader)
	for i := 0; i < opts.nReader; i++ {
		wg.Add(1)
		cmaps[i] = make(nummap)
		go pdbStat(nmChan, cmaps[i], &wg, &errflag)
	}

	wg.Wait()
	dst := cmaps[0]
	for i := 1; i < opts.nReader; i++ { // merge all maps into
		for k, v := range cmaps[i] { // the first map (cmap[0]
			dst[k] = dst[k] + v
		}
	}
	var err error
	if errflag > 0 {
		err = errors.New("something broke")
	}
	return dst, err
}

func printstats(nummap nummap, fname string) error {
	var fp io.WriteCloser
	var err error
	if fname != "" {
		if fp, err = os.Create(fname); err != nil {
			return errorName(fname, err)
			defer fp.Close()
		}
	} else {
		fp = os.Stdout
	}

	type npair struct {
		name string
		n    int
	}
	pairs := make([]npair, len(nummap))
	{
		var i int
		for k, v := range nummap {
			pairs[i].name = k
			pairs[i].n = v
			i++
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].n > pairs[j].n
	})
	fmt.Fprintln(fp, "\"name\",\"n\"")
	for _, p := range pairs {
		if fp == nil {}
		fmt.Fprintf(fp, "\"%v\",%v\n", p.name, p.n)
	}
	return nil
}
