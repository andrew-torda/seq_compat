#!/bin/sh
#$ -cwd
#$ -N test_parser
exe=./bigread.testing$$

gccgo=/work/torda/no_backup/gcc/bin/go
a=/work/torda/no_backup/gcc/lib64

date

LD_LIBRARY_PATH=$a $gccgo build -gccgoflags -O2 bigread.go && mv bigread bigread.gcc || exit 1
go build bigread.go && mv bigread bigread.goc || exit 1

d_arg=150
r_arg=3

for i in 1 2 3 ; do
    /usr/bin/time bigread.goc  -d $d_arg -r $r_arg > /dev/null
done

for i in 1 2 3 ; do
    LD_LIBRARY_PATH=$a /usr/bin/time bigread.gcc   -d $d_arg -r $r_arg > /dev/null
done




