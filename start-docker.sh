#!/bin/sh

docker-compose down || true
docker build -t securitybunker/databunker:latest .
docker-compose up -d
