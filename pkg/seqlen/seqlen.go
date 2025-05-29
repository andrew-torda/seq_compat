// 15 May 2025
// We are given a multiple sequence alignment.
// For each sequence, get the comment, species name and write send
// create a file for a spreadsheet.
// This will have the gene_id, the species name and the length
// of the sequence without gaps.

package seqlen

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/andrew-torda/seq_compat/pkg/seq"
)

const (
	ok byte = iota
	unhappy
	fatal
	stopping
)

type CmdArgs struct {
	InSeqFname  string // Read sequences from here
	OutCntFname string // Write counts (sequence lengths) to here
	OutSeqFname string // Optionally write sequences without gaps here
	IgnrSeqLen  bool   // Do not worry about sequences having different lengths
}

// nameSpecies gets the name and species from a comment
func nameSpecies(cmmt string) (name string, species string) {
	t := strings.Fields(cmmt)
	name = t[0]
	species = ""
	_, after, found1 := strings.Cut(cmmt, "[")
	if found1 {
		before, _, found2 := strings.Cut(after, "]")
		if found2 {
			species = before
		}
	}
	return name, species
}
func breaker() {}
func nothing ( x interface{}){}

// writeLen reads a sequence file, visits each sequence in turn
// and writes the length (and species) to an output (csv) file.
func writeLen(cmdArgs CmdArgs, outCntFile io.Writer) error {

	s_opts := &seq.Options{RmvGapsRd: true}
	if cmdArgs.IgnrSeqLen == true {
		s_opts.DiffLenSeq = true
	}
	seqgrp, err := seq.Readfile(cmdArgs.InSeqFname, s_opts)
	if err != nil {
		return fmt.Errorf("Fail reading sequences: %w", err)
	}

	csvDst := csv.NewWriter(outCntFile)

	for _, seq := range seqgrp.SeqSlc() {
		len := strconv.Itoa(len(seq.GetSeq()))
		cmmt := seq.Cmmt()
		name, species := nameSpecies(cmmt)
		s3 := [3]string{name, len, species}
		ss := s3[:]

		if err := csvDst.Write(ss); err != nil {
			return err
		}
	}

	return nil
}

// Mymain sets up files for reading and writing
func Mymain(cmdArgs CmdArgs) error {
	fmt.Println("main inseq", cmdArgs.InSeqFname, "outcnt",
		cmdArgs.OutCntFname, "outseq ", cmdArgs.OutSeqFname)
	var wrtNewSeqs bool // Should we write sequences without gaps
	if cmdArgs.OutSeqFname != "" {
		wrtNewSeqs = true
		return errors.New("Function to write new files not implemented. Try again later")
	}
	// inconsistent .. we open the writer here and the read
	// in the looping function
	var outCntFile io.WriteCloser
	if cmdArgs.OutCntFname == "-" {
		outCntFile = os.Stdout
	} else {
		var err error
		outCntFile, err = os.Create(cmdArgs.OutCntFname)
		if err != nil {
			return fmt.Errorf("output count file: %w", err)
		}
		defer outCntFile.Close()
	}

	if wrtNewSeqs {
		return errors.New("No way to get here")
	}

	if err := writeLen(cmdArgs, outCntFile); err != nil {
		return err
	}

	return nil
}
