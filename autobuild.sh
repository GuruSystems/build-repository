#!/bin/sh
# go insists on absolute path.
export GOPATH=`pwd`
echo "GOPATH=$GOPATH"

buildall() {
mkdir -p $GOBIN
MYSRC=src/golang.conradwood.net/build-repo/
( cd ${MYSRC} && make proto ) || exit 10
( cd ${MYSRC} && make client ) || exit 10
( cd ${MYSRC} && make server ) || exit 10
cp -rvf ${MYSRC}/proto dist/
}

export GOBIN=`pwd`/dist/i386
buildall
export GOBIN=`pwd`/dist/amd64
export GOOS=linux
export GOARCH=amd64
buildall

cp -v dist/i386/build-repo-client /usr/local/bin/ || exit 10
build-repo-client -branch=${GIT_BRANCH} -build=${BUILD_NUMBER} -commitid=${COMMIT_ID} -commitmsg="commit msg unknown" -repository=${PROJECT_NAME} -server_addr=buildrepo:5004 

exit 0
