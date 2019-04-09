#!/usr/bin/env bash
set -u

base=$(dirname $0)

j2=$(which j2)
if [ $? -ne 0 ]; then
    echo "j2 not found, install with pip"
    exit 1
fi

set -e
./${base}/generate-list.sh | j2 --format=json ${base}/.goreleaser.yml.j2 > .goreleaser.yml

