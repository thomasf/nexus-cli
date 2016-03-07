#!/bin/bash

set -e

release-build() {
  local tag=$1
  rm -rf build
  mkdir -p build
  for os in linux darwin windows; do
    for arch in amd64; do
      local fname=build/nexus-cli-${os}-${arch}-$tag
      [ "$os" == "windows" ] && fname="${fname}.exe"
      echo "building $fname"
      GOOS=${os} GOARCH=${arch} go build -o ${fname} main.go
    done
  done
  gzip build/*
}

tests() {
  echo $DRONE_TAG
  go test
}

case ${1} in
  build|tests)
    shift
    $1
    ;;
  runci)
    tests
    [[ -z $DRONE_TAG ]] || release-build $DRONE_TAG
    ;;
  *)
    echo "Nothing to do for ${1}"
    exit 1
    ;;
esac
