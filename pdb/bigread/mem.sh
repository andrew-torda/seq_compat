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
go build bigread.go
/usr/bin/time ./bigread -r 5 -d 300 -m mem.prof
go tool pprof -sample_index=alloc_space ./bigread mem.prof 
