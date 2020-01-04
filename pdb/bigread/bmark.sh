#!/bin/sh
#$ -q wap.q -l hostname=basel
#$ -cwd
##$ -pe openmpi_pe 6
#$ -N time_io
echo uname -a
# Warmup
./bigread -d 50 -r 8 2>/dev/null > /dev/null

for nt in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16  1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 ; do
    echo -n "$nt  "
    /usr/bin/time -f 'real %e user %U sys %S\n' ./bigread -d 0 -r $nt > /dev/null
done
