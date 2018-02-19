#!/bin/sh
# go insists on absolute path.
export GOPATH=`pwd`
echo "GOPATH=$GOPATH"

rm -rf dist

buildall() {
    echo "Building ${GOOS}/${GOARCH}"
    export GOBIN=$GOPATH/dist/${GOOS}/${GOARCH}/
    mkdir -p $GOBIN
    mkdir -p dist/${GOOS}/${GOARCH}/ || exit 10
    ( cd client && go build -o ../dist/${GOOS}/${GOARCH}/build-repo-client ) || exit 10
    ( cd server && go build -o ../dist/${GOOS}/${GOARCH}/build-repo-server) || exit 10
}

GOBIN=GOPATH/dist/proto

#========= build linux
export GOOS=linux
export GOARCH=amd64
buildall
mkdir -p dist/${GOOS}/${GOARCH}/ || exit 10
cp -p client/build-repo* dist/${GOOS}/${GOARCH}/ || exit 10

#========= build mac
export GOOS=darwin
export GOARCH=386
buildall

cp -v dist/linux/amd64/build-repo-client /usr/local/bin/ || exit 10

echo "Done OK"

build-repo-client -branch=${GIT_BRANCH} -build=${BUILD_NUMBER} -commitid=${COMMIT_ID} -commitmsg="commit msg unknown" -repository=${PROJECT_NAME} -server_addr=buildrepo:5004



exit 0
