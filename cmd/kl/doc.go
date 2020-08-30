// 30 Aug 2020
/*

kl compares two multiple sequence alignments and calculates the Kullbach-Leibler distance at each site, as well as the cosine similarity and the entropy for each alignment.

Results are written in csv format.

Usage:
 kl [options] file1.fa file2.fa

Flags:
  -f N
    	On output, we number each site starting from 1, but we can add
    	an offset of N to each value. It can be negative.
  -g	Gaps are a valid symbol. Without this option, gaps are ignored
  		in calculations.
  -n N
    	Treat the sequences as having N symbols. Without this, the
    	code will try to guess if we have nucleotides (4 symbols) or
    	proteins (20 symbols).

  -o filename
    	Write output to filename. If not give, numbers are written to
    	standard output

One should probably pick a filename that ends in  ".csv". This is not
enforced.

The columns in the output have labels which should be read by any plotting program that understands csv format. 

The two input files must be from multiple sequence alignments. The sequences within each file have to be of the same length, including gaps. The sequences in the two files have to be of the same length.

The Kullbach-Leibler distance, (kl= sum (p_a * log (p_a/q_a))) is fundamentally asymmetric, so we calculate the results using one file as the p distribution and also as the q distribution. If p_a or q_a is zero for some amino acid/nucleotide type, there is no correct behaviour. We take a pessimistic view. If you have N sequences and symbol "a" does not appear in one distribution, we take its frequency to be 1/(N+1). That is, if you have a 100 sequences and you do not see type X, we say the probability of X occurring is 1/101. The rationale is that this is the best estimate. If you have 10000 sequences and do not see symbol X, you can say that its probability is 1/10001. If you have 10 sequences, you cannot be so sure, so its probability becomes 1/11.

*/
package main
