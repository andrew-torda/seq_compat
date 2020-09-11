// 20 April 2020

package seq_test

import (
	"fmt"
	"log"
	"os"

	. "github.com/andrew-torda/seq_compat/pkg/seq"
)

// wrt_t5 writes testseqs5 to a file and returns the filename and an
// error.
var set1 = `>s1
ACDaae
>s2
CCD-af
> s3
CCQaag`

var set2 = `> reference sequence
X-BAAA-
> s2
-DBAA--
> s3
-GBA---
> s4
-JB----`

var sets = []string{set1, set2}

func ExampleRawCount() {
	s_opts := &Options{KeepGapsRd: true, DryRun: true}

	for _, s := range sets {
		f_tmp, err := wrtTmp(s)
		if err != nil {
			log.Fatal("writing testseq")
		}
		defer os.Remove(f_tmp)
		if seqgrp, err := Readfile(f_tmp, s_opts); err != nil {
			log.Fatal(err)
		} else {
			seqgrp.Upper()
			seqgrp.PrintFreqs("%5.0f")
		}
		fmt.Println()
	}
	// Output:
	//-     0    0    0    1    0    0
	//A     1    0    0    2    3    0
	//C     2    3    0    0    0    0
	//D     0    0    2    0    0    0
	//E     0    0    0    0    0    1
	//F     0    0    0    0    0    1
	//G     0    0    0    0    0    1
	//Q     0    0    1    0    0    0
	//
	//-     3    1    0    1    2    3    4
	//A     0    0    0    3    2    1    0
	//B     0    0    4    0    0    0    0
	//D     0    1    0    0    0    0    0
	//G     0    1    0    0    0    0    0
	//J     0    1    0    0    0    0    0
	//X     1    0    0    0    0    0    0

}

func scaledcheck(gapsAreChar bool) {
	s_opts := &Options{KeepGapsRd: true, DryRun: true}
	inner := func(f_tmp string, gapsAreChar bool) {
		if seqgrp, err := Readfile(f_tmp, s_opts); err != nil {
			log.Fatal(err)
		} else {
			seqgrp.Upper()
			seqgrp.UsageFrac(gapsAreChar)
			seqgrp.PrintFreqs("%6.2f")
			fmt.Println()
		}
	}
	for _, s := range sets {
		f_tmp, err := wrtTmp(s)
		if err != nil {
			log.Fatal("writing testseq")
		}
		defer os.Remove(f_tmp)
		inner(f_tmp, gapsAreChar)
	}
}

func ExampleScaledGapTrue() {
	scaledcheck(true)
	// Output:
	// 	-   0.00  0.00  0.00  0.33  0.00  0.00
	// A   0.33  0.00  0.00  0.67  1.00  0.00
	// C   0.67  1.00  0.00  0.00  0.00  0.00
	// D   0.00  0.00  0.67  0.00  0.00  0.00
	// E   0.00  0.00  0.00  0.00  0.00  0.33
	// F   0.00  0.00  0.00  0.00  0.00  0.33
	// G   0.00  0.00  0.00  0.00  0.00  0.33
	// Q   0.00  0.00  0.33  0.00  0.00  0.00

	// -   0.75  0.25  0.00  0.25  0.50  0.75  1.00
	// A   0.00  0.00  0.00  0.75  0.50  0.25  0.00
	// B   0.00  0.00  1.00  0.00  0.00  0.00  0.00
	// D   0.00  0.25  0.00  0.00  0.00  0.00  0.00
	// G   0.00  0.25  0.00  0.00  0.00  0.00  0.00
	// J   0.00  0.25  0.00  0.00  0.00  0.00  0.00
	// X   0.25  0.00  0.00  0.00  0.00  0.00  0.00
}

func ExampleScaledGapFalse() {
	scaledcheck(false)
	// Output:
	// 	-   0.00  0.00  0.00  0.33  0.00  0.00
	// A   0.33  0.00  0.00  1.00  1.00  0.00
	// C   0.67  1.00  0.00  0.00  0.00  0.00
	// D   0.00  0.00  0.67  0.00  0.00  0.00
	// E   0.00  0.00  0.00  0.00  0.00  0.33
	// F   0.00  0.00  0.00  0.00  0.00  0.33
	// G   0.00  0.00  0.00  0.00  0.00  0.33
	// Q   0.00  0.00  0.33  0.00  0.00  0.00

	// -   0.75  0.25  0.00  0.25  0.50  0.75  1.00
	// A   0.00  0.00  0.00  1.00  1.00  1.00  0.00
	// B   0.00  0.00  1.00  0.00  0.00  0.00  0.00
	// D   0.00  0.33  0.00  0.00  0.00  0.00  0.00
	// G   0.00  0.33  0.00  0.00  0.00  0.00  0.00
	// J   0.00  0.33  0.00  0.00  0.00  0.00  0.00
	// X   1.00  0.00  0.00  0.00  0.00  0.00  0.00
}
