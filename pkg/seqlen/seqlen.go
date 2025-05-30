// 15 May 2025
// We are given a multiple sequence alignment.
// For each sequence, get the comment, species name and write send
// create a file for a spreadsheet.
// This will have the gene_id, the species name and the length
// of the sequence without gaps.

package seqlen

import (
	"encoding/csv"
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
// Name is just the first word in the comment.
// Species is any text between [ and ].
func nameSpecies(cmmt string) (name string, species string) {
	name = strings.Fields(cmmt) [0] // Get the first word
	if _, after, found1 := strings.Cut(cmmt, "["); found1 {
		if before, _, found2 := strings.Cut(after, "]"); found2 {
			species = before
		}
	}
	return name, species
}

// writeLen reads a sequence file, visits each sequence in turn
// and writes the length (and species) to an output (csv) file.
func writeLen(cmdArgs CmdArgs, outCntFile io.Writer) error {
	s_opts := &seq.Options{RmvGapsRd: true, ZeroLenOK: true}
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

		if err := csvDst.Write(s3[:]); err != nil {
			return err
		}
	}
	csvDst.Flush()

	if cmdArgs.OutSeqFname != ""{
		osf:= cmdArgs.OutSeqFname // readability
		if err := seq.WriteToF(osf, seqgrp.SeqSlc(), s_opts); err != nil{
			e := fmt.Errorf("csv file OK, but error writing seqs: %w", err)
			return e
		}
	}
	return nil
}

// Mymain sets up files for reading and writing.
// Must do a bit of work so it can also read from standard input.
func Mymain(cmdArgs CmdArgs) error {
	var outCntFile io.WriteCloser
	if cmdArgs.OutCntFname == "-" { // Sort out the outputfile
		outCntFile = os.Stdout
	} else {
		var err error
		outCntFile, err = os.Create(cmdArgs.OutCntFname)
		if err != nil {
			return fmt.Errorf("output count file: %w", err)
		}
		defer outCntFile.Close()
	}

	if err := writeLen(cmdArgs, outCntFile); err != nil {
		return err
	}

	return nil
}
