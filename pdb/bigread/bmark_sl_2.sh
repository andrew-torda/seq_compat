#!/bin/sh
#$ -q hpc.q
#$ -cwd
##$ -pe openmpi_pe 6
#$ -N time_slice_siz
uname -a
date

export GOBIN=/home/torda/go/bin
#GOARCH=amd64
export GOROOT=/work/torda/no_backup/go
#GOOS=linux
export GOPATH=/home/torda/go

go clean
go build bigread.go
# Warmup
sl_siz=50 ./bigread -r 6 2>/dev/null > /dev/null


for siz in 25 30 40 45 50 55 60 62 64 70 100 25 30 40 45 50 55 60 100 40 41 42 43 44 45 46 47 48 49 50 40 41 42 43 44 45 46 47 48 49 50 ; do
    echo -n slice_size $siz
    sl_siz=$siz    /usr/bin/time -f ' real %e user %U sys %S RSS %M\n' ./bigread  -r 6 > /dev/null
done
date
