#!/bin/sh

VERSION=$(cat ./version.txt)
docker build -t securitybunker/databunker:$VERSION .
docker tag securitybunker/databunker:$VERSION securitybunker/databunker:latest
docker-compose -f docker-compose-pgsql.yml down || true
docker-compose -f docker-compose-pgsql.yml up -d
