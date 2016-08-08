#!/bin/bash
#
# Generate n lines of pseudo random output to stdout and stderr.
#
# It accepts two arguments:
#     N - the number of lines to generate
#     M - the size of each line
#
N=${1-1000}
M=${2-72}
for ((i=0; i<N; i++)) ; do
    Line=$(LC_CTYPE=C tr -dc A-Za-z0-9 </dev/urandom | head -c $M)
    if (( (i % 2) == 0 )) ; then
        printf "%06d %s\n" $i $Line
    else
        printf "%06d %s\n" $i $Line 1>&2
    fi
done

