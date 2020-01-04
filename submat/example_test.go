package submat_test

import (
	"github.com/andrew-torda/goutil/submat"
	"github.com/andrew-torda/goutil/gotoh"
	"fmt"
)


func Example_scoreSeqs() {
	seqs := []string{"acdefgacdefg", "cdefgacsfg", "cdefgactg", "cdefgacwg"}
	pnltys := gotoh.Pnlty {Open: 2, Wdn: 2}
	substMat, err := submat.Read("blosum62.txt")
	if  err != nil {
		fmt.Print(err)
	}
	al_details := gotoh.Al_score {
		Pnlty: pnltys,
		Al_type: gotoh.Local,
	}
	for i, s := range seqs {
		for j := i +1 ; j < len(seqs) ; j++ {
			t := seqs[j]
			scr_mat := substMat.ScoreSeqs([]byte(s), []byte(t))
			pairlist, a_scr := gotoh.Align (scr_mat, &al_details)
			gotoh.PrintSeqDebug(true, pairlist, []byte(s), []byte(t), gotoh.Global)
			fmt.Println ("score", a_scr)
		}
	}
}
