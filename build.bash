#!/bin/bash -xe

GOOS=linux GOARCH=amd64 gb build
GOOS=windows GOARCH=amd64 gb build
GOOS=darwin GOARCH=amd64 gb build

