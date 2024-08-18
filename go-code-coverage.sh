#!/bin/bash

cd src
go test -coverprofile=/tmp/coverage.out
echo
echo "Test summary"
go tool cover -func=/tmp/coverage.out
echo
go tool cover -html=/tmp/coverage.out
