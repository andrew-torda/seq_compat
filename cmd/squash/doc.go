// 20 April 2020
// 27 april 2020

/*
Squash removes columns from a multiple sequence alignment

We take a multiple sequence alignment and some specified sequence.
From the multiple sequence alignment remove any columns which
correspond to a gap in the specified sequence.

Usage:
	squash name [input] [output]


If no output file is given, stdout will be used.
If no input file is given, stdin will be used.

The name argument is tricky. It is a string, so it will have to be quoted
on the command line.
It must be contained within the comment of a sequence, so
 'blah blah [homo sapiens]'
will look for 'blah blah [homo sapiens]' within the sequences. It will
stop looking at the first match it finds.
Be careful. If there is a sequence that starts
 > blah blah [homo sapiens]
and there is another that starts
 > blah blah foo foo [monkey]
and you search for "blah blah", it will return the first sequence it
finds, even if you want the monkey sequence.
There is more.
If the string starts and ends with an integer like "1" it will take
this as the number of the sequence to be removed.
Very often, the first sequence is the reference, so you would say

 1

instead of a sequence name.


*/
package main


