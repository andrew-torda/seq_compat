// 31 July 2020

/*

Randseq is for making random sequences for testing the code.
Usage:
	randseq [options] fname nseq length
will generate nseq sequences of length length and write them to fname.

Flags:
	-g
		no gaps in the output sequences
	-e
		provoke errors. Sequences will not be the same length. This is an error
		for programs like entropy.
	-r
		random number seed

We are most interested in benchmarking and parsing, so the content is not so important.
The only question that comes up is white space and gaps.
Whitespace should generally be unpredictable, so we generate funny cases.
There is a flag (-g) which tells us whether we have gaps.
Most programs we use expect sequences to be the same length (after gap removal),
so we can provoke errors and have different lengths.

There are two main entry points.
 1. Write to a file
 2. Write to a string.

*/
package main
