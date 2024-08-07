#!/bin/sh

docker compose -f docker-compose-pgsql.yml down || true
docker compose -f docker-compose-pgsql.yml up -d
