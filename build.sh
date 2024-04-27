#!/bin/bash

set -x

VERSION=$(cat ./version.txt)

if [[ ! -x "~/go/bin/packr" && ! -x "packr" ]]; then
  go install github.com/gobuffalo/packr/packr@latest
fi

cd src
go get -d -v
if [ -x "~/go/bin/packr" ]; then
  ~/go/bin/packr
elif [ -x "packr" ]; then
  packr
fi

go build -v -ldflags="-s -w -X main.version=${VERSION}" -o ../databunker

if [ -x "~/go/bin/packr" ]; then
  ~/go/bin/packr clean
elif [ -x "packr" ]; then
  packr clean
fi

cd ..
