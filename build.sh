#!/bin/sh

set -x

VERSION=$(cat ./version.txt)
HASH=$(git rev-parse --short=12 HEAD)
TIMESTAMP=$(date +%Y%m%d%H%M%S)  # Current timestamp
FULL_VERSION=$VERSION-$TIMESTAMP-$HASH

cd src
go get -d -v
go build -v -ldflags="-s -w -X main.version=$FULL_VERSION" -o ../databunker
cd ..
