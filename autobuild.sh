#!/bin/sh
# go insists on absolute path.
export GOPATH=`pwd`
echo "GOPATH=$GOPATH"

buildall() {
    GOBIN=GOPATH/dist/${GOOS}/${GOARCH}/
    mkdir -p $GOBIN
    MYSRC=src/golang.conradwood.net/build-repo/
    ( cd ${MYSRC} && make proto ) || exit 10
    ( cd ${MYSRC} && make client ) || exit 10
    ( cd ${MYSRC} && make server ) || exit 10
    cp -rvf ${MYSRC}/proto dist/
}

GOBIN=GOPATH/dist/proto
( cd ${MYSRC} && make proto ) || exit 10

#========= build linux
export GOOS=linux
export GOARCH=amd64
buildall

#========= build mac
export GOOS=darwin
export GOARCH=386
buildall


cp -v dist/amd64/build-repo-client /usr/local/bin/ || exit 10
build-repo-client -branch=${GIT_BRANCH} -build=${BUILD_NUMBER} -commitid=${COMMIT_ID} -commitmsg="commit msg unknown" -repository=${PROJECT_NAME} -server_addr=buildrepo:5004 

exit 0
