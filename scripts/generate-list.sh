#!/usr/bin/env bash
set -eu

base=$(dirname $0)/..


echo -n '{"builds": ['
c=0
find ${base} -name "main.go" | while read src; do
    src=$(realpath --relative-to=${base} ${src})
    fname=$(echo ${src} | awk -F/ '{print $1"-"$2}')
    dir=$(echo ${src} | awk -F/ '{print $1"/"$2}')
    if [ $c -ne 0 ]; then
	     echo ","
    else
	     echo ""
    fi
    echo -n '{"src": "./'${dir}/'", "name":"'${fname}'"}'
    c=$((c + 1))
done
echo ""
echo "]}"
