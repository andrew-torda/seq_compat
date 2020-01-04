#!/bin/sh
#$ -q hpc.q
#$ -cwd
##$ -pe openmpi_pe 6
#$ -N time_slice_siz
echo uname -a
# Warmup
sl_siz=50 ./bigread -d 50 -r 8 2>/dev/null > /dev/null

for siz in 1 5 10 20 50 100 200 500 ; do
    echo slice size $siz
    sl_siz=$siz    /usr/bin/time -f 'real %e user %U sys %S RSS %M\n' ./bigread  -r 6 > /dev/null
done
