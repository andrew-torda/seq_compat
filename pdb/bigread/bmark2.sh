#!/bin/sh
#$ -q wap.q
#$ -l hostname=herne
#$ -cwd
##$ -pe openmpi_pe 6
#$ -N time_io
uname -a
export GOBIN=/home/torda/go/bin
export GOROOT=/work/torda/no_backup/go
export GOPATH=/home/torda/go

go clean
goexe="go run bigread.go"
# Warmup
$goexe -d 50 -r 5 2>/dev/null > /dev/null

for nt in 5 6  ; do
    echo -n "$nt  "
    /usr/bin/time -f 'real %e user %U sys %S\n' $goexe -d 0 -r $nt > /dev/null
done
