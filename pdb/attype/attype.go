// attype will read lots of pdb files and collect all the different atom types it finds.

package main

import (
	"errors"
	"fmt"
	"flag"
	"os"
	"path"
)

const (
	parentPath   = "/work/public/no_backup/pdb/data/structures/divided/mmCIF/"
	exit_success = 0
	exit_failure = 1
	maxPDBerr    = 5 // The maximum number of broken PDB files, before a reader gives up
)

// A fileRes gives us the name of the next file, but also includes
// room for an error
type fileRes struct {
	name string
	err  error
}

// errorName sticks a problem causing filename on an error message
func errorName(fname string, e error) error {
	s := "Working on \"" + fname + "\": " + e.Error()
	return errors.New(s)
}

// nextPfile looks in the PDB directories starting from parentPath,
// visiting each subdirectory and sending the names down the nmChan
// channel.
// We stop after MaxFile files, but if maxFile <= 0, we just read
// until there are no more.
func nextPfile(nmChan chan fileRes, parentPath string, maxFile int) {
	defer close(nmChan)
	dontStop := true 
	if maxFile > 0 {
		dontStop = false
	}
	topdir, err := os.Open(parentPath)
	if err != nil {
		nmChan <- fileRes{"", err}
		return
	}
	var dlist []string
	if dlist, err = topdir.Readdirnames(0); err != nil {
		nmChan <- fileRes{"", errorName(parentPath, err)}
		return
	}

	var namebuf []string
	var dname string
	for ndone := 0; ndone < maxFile || dontStop; {
		for len(namebuf) == 0 && len(dlist) > 0 {
			dname, dlist = dlist[0], dlist[1:]
			dname = path.Join(parentPath, dname)
			if cdir, err := os.Open(dname); err != nil {
				nmChan <- fileRes{"", err}
				return
			} else {
				if namebuf, err = cdir.Readdirnames(0); err != nil {
					nmChan <- fileRes{"", errorName(dname, err)}
					return
				}
				cdir.Close()
			}
		}
		if len(namebuf) == 0 { // We have run out of directories to look
			return //             in, so just give up.
		}
		var t string
		t, namebuf = namebuf[0], namebuf[1:]
		t = path.Join (dname, t)
		nmChan <- fileRes{t, nil}
		ndone++
	}
}

type opts struct {
	nReader int
	maxFile int
	outFname string }
func doFlags (opts *opts) {
	flag.IntVar(&opts.nReader, "r", opts.nReader, "num reader threads")
	flag.IntVar(&opts.maxFile, "f", opts.maxFile, "max num files to read")
	flag.StringVar(&opts.outFname, "o", opts.outFname, "output filename instead of stdout")
	flag.Parse()
}	

func mymain() int {
	opts := opts { nReader: 3, maxFile: 3, outFname: ""}
	doFlags (&opts)
	nmChan := make (chan fileRes)
	go nextPfile(nmChan, parentPath, opts.maxFile )
	if nummap, err := collectData (nmChan, opts) ; err != nil{
		fmt.Fprintln (os.Stderr, err)
		return exit_failure
	} else {
		printstats (nummap, opts.outFname)
	}
	return exit_success
}

func main() {
	os.Exit(mymain())
}
