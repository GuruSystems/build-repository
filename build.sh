#!/bin/sh
export GOPATH=`pwd`

[ -z $1 ] && echo "Arguments must be 1: OS, 2: ARCH, 3: TARGET" && exit 10

export BUILD_PATH=../dist/${1}/${2}/
mkdir -p ${BUILD_PATH} || exit 10
export BUILD_PATH="${BUILD_PATH}build-repo-${3}"

echo "Building ${BUILD_PATH}"

cd $3 || exit 10

go build -o $BUILD_PATH || exit 10
echo "Built OK"
