#!/bin/sh

VERSION=$(cat ./version.txt)
HASH=$(git rev-parse --short=12 HEAD)
TIMESTAMP=$(date +%Y%m%d%H%M%S)  # Current timestamp
FULL_VERSION=$VERSION-$TIMESTAMP-$HASH

if [ -x ~/go/bin/packr ]; then
  echo "Found ~/go/bin/packr"
elif [ -x "packr" ]; then
  echo "Fond packr"
else
  echo "Installing packr"
  go install github.com/gobuffalo/packr/packr@latest
fi

cd src
go get -v

if [ -x ~/go/bin/packr ]; then
  ~/go/bin/packr
else
  packr
fi

CGO_ENABLED=0 go build -v -ldflags="-s -w -X main.version=$FULL_VERSION" -o ../databunker

if [ -x ~/go/bin/packr ]; then
  ~/go/bin/packr clean
else
  packr clean
fi

cd ..
