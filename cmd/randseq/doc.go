// 31 July 2020

/*

Randseq is for making random sequences for testing the code.
We are most interested in benchmarking and parsing, so the content is not so important.
The only question that comes up is white space and gaps.
Whitespace should generally be unpredictable, so we encounter funny cases.
There is a flag which tells us whether we have gaps.
We should add a flag to generate sequences of different lengths, so as to
provoke errors.

There are two main entry points.
 1. Write to a file
 2. Write to a string.

*/
package main
