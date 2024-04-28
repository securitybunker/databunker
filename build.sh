#!/bin/sh

set -x

VERSION=$(cat ./version.txt)
HASH=$(git rev-parse --short=12 HEAD)
TIMESTAMP=$(date +%Y%m%d%H%M%S)  # Current timestamp
FULL_VERSION=$VERSION-$TIMESTAMP-$HASH

if [ -x "~/go/bin/packr" ]; then
  echo "Found ~/go/bin/packr"
elif [ -x "packr" ]; then
  echo "Fond packr"
else
  go install github.com/gobuffalo/packr/packr@latest
fi

cd src
go get -d -v

if [ -x "~/go/bin/packr" ]; then
  ~/go/bin/packr
elif [ -x "packr" ]; then
  packr
fi

go build -v -ldflags="-s -w -X main.version=$FULL_VERSION" -o ../databunker

if [ -x "~/go/bin/packr" ]; then
  ~/go/bin/packr clean
elif [ -x "packr" ]; then
  packr clean
fi

cd ..
