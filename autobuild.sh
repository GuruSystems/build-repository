#!/bin/sh

echo "Buidling all applications..."

rm -rf dist

sh build.sh linux amd64 client || (echo "Failed..."; exit 10)
sh build.sh linux amd64 server || (echo "Failed..."; exit 10)
sh build.sh darwin 386 client || (echo "Failed..."; exit 10)
sh build.sh darwin 386 server || (echo "Failed..."; exit 10)

# If Mac arguement is given use the darwin/386 build-repo binary
export THISOS="linux/amd64"
[ -z $1 ] || export THISOS="darwin/386"

cp -v dist/${THISOS}/build-repo-client /usr/local/bin/ || exit 10

echo "Sending to ${THISOS} build-repo..."

[ -z "${BUILD_NUMBER}" ] && export BUILD_NUMBER=${CI_PIPELINE_ID}
[ -z "${PROJECT_NAME}" ] && export PROJECT_NAME=${CI_PROJECT_NAME}
[ -z "${COMMIT_ID}" ] && export COMMIT_ID=${CI_COMMIT_SHA}
[ -z "${GIT_BRANCH}" ] && export GIT_BRANCH=${CI_COMMIT_REF_NAME}

build-repo-client -branch=${GIT_BRANCH} -build=${BUILD_NUMBER} -commitid=${COMMIT_ID} -commitmsg="commit msg unknown" -repository=${PROJECT_NAME} -server_addr=buildrepo:5004

exit 0
