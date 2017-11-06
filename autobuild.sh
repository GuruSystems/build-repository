#!/bin/sh
# go insists on absolute path.
export GOBIN=`pwd`/dist
export GOPATH=`pwd`
echo "GOPATH=$GOPATH"
mkdir $GOBIN
MYSRC=src/golang.conradwood.net/build-repo/
( cd ${MYSRC} && make proto ) || exit 10
( cd ${MYSRC} && make client ) || exit 10
( cd ${MYSRC} && make server ) || exit 10
cp -rvf ${MYSRC}/proto dist/

cp dist/build-repo-client /usr/local/bin/
exit 0
