#!/bin/bash -xe

export GOROOT=$(go env GOROOT)

GOOS=linux GOARCH=amd64 gb build
GOOS=windows GOARCH=amd64 gb build
GOOS=darwin GOARCH=amd64 gb build

