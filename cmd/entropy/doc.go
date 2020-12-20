// 27 april 2020

/*

Entropy calculates sequence entropy from a multiple sequence alignment.
If given a reference sequence, it will also calculate the probability of finding
that sequence's residue at a particular position.

Given no output filename, it write to standard output.
Gaps are normally ignored. Calculations are done using only the
residues/bases in each column, but there is an option to treat gaps as a valid character.
The code guesses whether you have DNA or protein by looking at the symbols. This can be fooled, so you can also specify how many symbols should be used. This only affects the base of the logarithm in the calculation. If you have a DNA sequence, but there are some "X" characters, the code will think it has an alphabet of size 5. You would then have to tell it to use four symbols.

The output is a csv file for plotting with some other program. It has a header line which programs like excel like, but gnuplot is less keen on. R's read.csv() has an option to tell it there is a header line.


Usage:
	entropy [flags] [input]

The flags are:
	-c chimera_attribute_file
		Write conservation data to a file in the format of an attribute file that chimera can read and use for coloring a structure.
	-f oFfset
		When creating output for plotting, we assume the first residue is numbered 1. This allows one to add an offset to be added or subtracted (if negative) to each number.
	-g
		Treat gaps as a valid character
	-n base
		Set the base for logarithms and override the guess. 20 for protein. 4 for DNA.
	-o Outfilename
		Output file name, instead of standard output
	-r reference
		Specify a reference sequence by give a string which will be searched
		for in the comment lines of the sequences

If you have a reference sequence, the compatibility of each base/residue will be calculated and printed out.

OUTPUT
Output is written to the standard output, which you probably do not want. Catch the output on the command line with a redirection or add a second filename which will be the output filename.

The format is .csv with quoted heading for the columns. It can be eaten with read.csv in R or imported straight into excel. Gnuplot also knows what to do with it.

TODO
One could add a mapping of input residues to output, so, for example, selenomethionine becomes methionine.
*/
package main
