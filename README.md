# Intro
This is a small set of packages for working with conservation in multiple sequence alignments.
It works under linux, windows and maybe on a mac. I don't have a mac

# Installation

Is go installed on your machine ? If so,
`go get github.com/andrew-torda/seq_compat`
If you then 
```
 go build ./...
 go test ./...
```
You will end up with an executable in each of the directories under `cmd`.
Doing
` go install ./...`
Will put these executables somewhere.
otherwise,
 git clone https://github.com/andrew-torda/seq_compat.git

## Documentation
From the top level, you can type
 go doc entropy
or
 go doc kl
or
 go doc squash
to see a brief description. Less interesting, is `go doc randseq`.

# programs
## kl
Given two files, calculate the per-site Kullbach-Leibler distance, as well as the cosine similarity.

## entropy
Calculate the per-site entropy in a multiple sequence alignment. Write it in .csv format for plotting in gnuplot/R/excel/whatever.

## squash
Takes an input multiple sequence alignment and a reference sequence. It produced the multiple sequence alignment, but with only the columns present where the reference sequence has a character and not a gap.

## randseq
Generates random, fasta-formatted sequences. It is only useful for testing. The sequences are pleasantly awful with white space all over place.

## numseq
This is also only used in testing. It is really only a wrapper for the package in the `pkg/numseq` directory. This executable does not do anything cleverer than 
 grep '>' | wc -l
although it is much faster.

# Examples

coming

# Structure
There are a few commands. These live under `cmd`. They are very short and quickly drop into a package which lives under `pkg`.

# To Do

Add an option for chimera output.

# Implementation
Sequences are read by the `seq` package. This is used by the other programs. `seq` has a `seq` structure and a `seqgrp` structure. `seq`s have a comment (utf-8 strings) and a sequence (a set of ascii bytes).



# Regrets

## Precision
This is not a regret. All the calculations are done in single precision. That is fine and accurate enough for us. It does mean that the code is full of `float32` casts.

## Memory layout
I suspect the code would be faster if the residues were stored column-major, not row-major.
