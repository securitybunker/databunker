#!/bin/sh

VERSION=$(cat ./version.txt)
docker build -t securitybunker/databunker:$VERSION .
docker tag securitybunker/databunker:$VERSION securitybunker/databunker:latest
docker push securitybunker/databunker:$VERSION
docker push securitybunker/databunker:latest
