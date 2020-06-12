package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"

	"github.com/andrew-torda/goutil/pdb"
	"github.com/andrew-torda/goutil/pdb/cmmn"
)

const parentPath = "/work/public/no_backup/pdb/data/structures/divided/mmCIF/"

const (
	nReaderDflt = 3   // Default number of reader threads
	maxDirDflt  = 200 // Read this many directories
	exitSuccess = 0
	exitFail    = 1
)

type fresult struct {
	nbyte int64
	nfile int
}

func breakerp(p interface{}) {}

// readpdb reads a file
func readpdb(fullpath string) (int, error) {
	fp, err := os.Open(fullpath)
	if err != nil {
		return 0, err
	}
	defer fp.Close()
	if p, err := pdb.ReadCoord(fullpath, cmmn.FileSrc, "testoutput"); err != nil {
		return 0, err
	} else {
		breakerp(p)
	}
	return 1, nil
	// else {
	//		jValid, jInvalid := pdb.NatomsTot (ch)
	//		b := filepath.Base(fullpath)
	//		fmt.Println (b, pdb.NChain(ch), jValid, jInvalid, pdb.ChainNames(ch))
	//	}

}

// readPdir takes the name of a directory (PDB) from a channel.
// It opens the directory and then calls readpdb on every file
// it finds.
func readPdir(ch <-chan string, res chan<- fresult,
	wg *sync.WaitGroup, parentpath string) {
	defer wg.Done()
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
			nbyte, err := readpdb(fullpath)
			if err != nil {
				fmt.Fprintln(os.Stderr, file, err)
			}
			fres.nbyte += int64(nbyte)
			fres.nfile++
		}
	}
	res <- fres
}

func mymain() (retval int) {
	var nReader int
	var maxDir int
	var brokenio bool
	var cpuprof, memprof, traceprof string
	flag.IntVar(&nReader, "r", nReaderDflt, "num reader threads")
	flag.IntVar(&maxDir, "d", maxDirDflt, "max num directories to read")
	flag.BoolVar(&brokenio, "b", false, "Use broken I/O for testing")
	flag.StringVar(&cpuprof, "c", "", "write cpuprofile to file")
	flag.StringVar(&memprof, "m", "", "write memprofile to file")
	flag.StringVar(&traceprof, "t", "", "write trace to file")
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

	c := make(chan string, 200)
	res := make(chan fresult)

	go func() {
		n := 0
		for _, file := range dirs {
			if maxDir == 0 || n < maxDir {
				c <- file
			}
			n++
		}
		close(c)
	}()

	if cpuprof != "" {
		fprof, err := os.Create(cpuprof)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return exitFail
		}
		if err := pprof.StartCPUProfile(fprof); err != nil {
			fmt.Println(err)
			return exitFail
		}
		defer fprof.Close()
		defer pprof.StopCPUProfile()
	}

	if traceprof != "" {
		tprof, err := os.Create(traceprof)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return exitFail
		}
		if err := trace.Start(tprof); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitFail
		}
		defer tprof.Close()
		defer trace.Stop()
	}

	var wg sync.WaitGroup
	for i := 0; i < nReader; i++ {
		wg.Add(1)
		go readPdir(c, res, &wg, parentPath)
	}
	var totalres fresult
	for i := 0; i < nReader; i++ {
		f := <-res
		fmt.Println("nbyte", f.nbyte, "nfile", f.nfile)
		totalres.nbyte += f.nbyte
		totalres.nfile += f.nfile
	}
	const mb = 1024 * 1024
	nbyte := float32(totalres.nbyte) / float32(mb)
	fmt.Printf("Totals nbyte %.2f Mb nfiles %d\n", nbyte, totalres.nfile)
	wg.Wait()
	if memprof != "" {
		runtime.GC()
		cprof, err := os.Create(memprof)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return exitFail
		}
		if err := pprof.WriteHeapProfile(cprof); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			return exitFail
		}
		cprof.Close()
	}

	return exitSuccess
}

func main() {
	os.Exit(mymain())
}
