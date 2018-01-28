#!/usr/bin/env bash

base=$(dirname $0)/..
version=$(cat ${base}/VERSION)

git tag -s v${version} -m "v${version}"
git push origin v${version}
