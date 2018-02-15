#!/bin/sh
export GOPATH=`pwd`
(cd server && go build)
