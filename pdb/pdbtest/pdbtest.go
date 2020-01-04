package main

import (
	"fmt"
	"os"
	"pdb"
	"pdb/cmmn"
	"runtime"
)

const (
	exitSuccess = 0
	exitFailure = 1
)

func mymain() (retval int) {
	const logFilename string = ""
	type fname struct {
		name  string
		stype byte
	}
	const cifDir = "/work/public/no_backup/pdb/data/structures/divided/mmCIF/"
	testfiles := []fname{
		{"2mzi", cmmn.HTTPSrc}, // anisotropic B-factors, hetatm mixed with atom records
		//		{"3j3q", mmcif.HTTPSrc},                    // Stupid monster structure
		{"3en1.cif", cmmn.FileSrc},
		{"2a9w.cif", cmmn.FileSrc},                // shorted version of next file
		{cifDir + "a9/2a9w.cif.gz", cmmn.FileSrc}, // error line 532
		{cifDir + "ad/2ad1.cif.gz", cmmn.FileSrc}, // unterm quote line 108
		{cifDir + "pt/5pti.cif.gz", cmmn.FileSrc}, // 20 models
		{"./four_nmr.cif", cmmn.FileSrc},
		{cifDir + "bg/6bgg.cif.gz", cmmn.FileSrc}, // non-standard linkages, more than one model
	}
	for _, s := range testfiles {
		if _, err := pdb.ReadCoord(s.name, s.stype, logFilename); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR on ", s.name, ": ", err)
		}
		runtime.GC()
	}

	return exitSuccess
}

func main() {
	os.Exit(mymain())
}
