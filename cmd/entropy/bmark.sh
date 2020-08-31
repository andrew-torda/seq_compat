# This is a reminder of the benchmarking, comparing the old perl code to
# the go version.
tfmt='%U %M'
newprog=~/g2/seq_compat/cmd/entropy/entropy
oldprog=~/bin/entropy.pl

junk=/home/torda/junk
inlist="$junk/2_2.fa $junk/5000_500.fa $junk/5000_1500.fa"

runprog() {
    progname=$1
    infile=$2
#   run it once to get everything cached
    $progname $infile  2>/dev/null >/dev/null
    for i in 1 2 3 4 5 6 7 8 9 10; do
        /usr/bin/time -f "$tfmt" $progname $infile > /dev/null
    done
}



for input in $inlist ; do
    for prog in $newprog $oldprog; do
        echo $prog $input
        runprog $prog $input
    done
done
