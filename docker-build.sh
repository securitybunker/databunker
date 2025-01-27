#!/bin/sh

VERSION=$(cat ./version.txt)
BUILDKIT_PROGRESS=plain docker build -t securitybunker/databunker:$VERSION .
docker tag securitybunker/databunker:$VERSION securitybunker/databunker:latest
