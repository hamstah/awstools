#!/usr/bin/env bash
set -e

find . -name "main.go" | while read f; do
    echo "Building ${f}"
    cd $(dirname ${f})
    CGO_ENABLED=0 go build -a -installsuffix cgo
    cd - > /dev/null
done
