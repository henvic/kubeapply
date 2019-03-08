#!/bin/bash

# WeDeploy CLI Tool static analysis scripts

set -euo pipefail
IFS=$'\n\t'

cd $(dirname $0)/..

echo "Checking for unchecked errors."
errcheck $(go list ./...)
echo "Linting code."
test -z "$(golint `go list ./...` | tee /dev/stderr)"
echo "Examining source code against code defect."
go vet $(go list ./...)
go vet -vettool=$(which shadow)
echo "Running staticcheck toolset https://staticcheck.io"
staticcheck ./...
echo "Checking if code contains security issues."
# TODO(henvic) fix/update gosec when available:
# G104: Doesn't understand _ assignments #270
# https://github.com/securego/gosec/issues/270
# Ignoring G104: Audit errors not checked for now.
gosec -quiet -exclude G104 --quiet ./...

go test $(go list ./... | grep -v /integration$) -race
