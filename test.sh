#!/bin/sh
[ -z $1 ] && (echo "Arguement must be target app name..."; exit 10)
export GOPATH=`pwd`
(cd $1 && go test -v)
