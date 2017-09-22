#!/usr/bin/env bash

chars=$(echo "0 1 2 3 4 5 6 7 8 9 a b c d e f" | cut -d' ' -f 1-)

if [ "$1" = "" ]; then
    mkdir ./data
    basedir="./data"
else
    basedir=$1
fi

for c1 in ${chars[@]}; do
    for c2 in ${chars[@]}; do
        mkdir $basedir/$c1$c2
    done
done
