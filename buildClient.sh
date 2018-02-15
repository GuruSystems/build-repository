#!/bin/sh
export GOPATH=`pwd`
(cd client && go build)
