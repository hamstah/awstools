#!/usr/bin/env bash
set -e

base=$(dirname $0)/..

mkdir -p ${base}/bin

echo "Building"
find ${base} -name "main.go" | while read src; do
    src=$(realpath --relative-to=${base} ${src})
    name=bin/$(echo ${src} | awk -F/ '{print $1"-"$2}')
    echo "  ${name}"
    CGO_ENABLED=0 go build -installsuffix cgo -o ${base}/${name} ${src}
    gpg --armor --detach-sig ${base}/${name}
done
