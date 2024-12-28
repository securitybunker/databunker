#!/bin/sh

VERSION=$(cat ./version.txt)
docker build --progress=plain -t securitybunker/databunker:$VERSION .
docker tag securitybunker/databunker:$VERSION securitybunker/databunker:latest
