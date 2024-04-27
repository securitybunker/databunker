#!/bin/sh

VERSION=$(cat ./version.txt)
docker build -t securitybunker/databunker:$VERSION --build-arg VERSION=$VERSION .
docker-compose -f docker-compose-pgsql.yml down || true
docker-compose -f docker-compose-pgsql.yml up -d
