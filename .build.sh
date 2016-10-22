#!/usr/bin/env bash

set -xe

cross_compile() {
  gox -arch="amd64" \
    -os="darwin linux windows" \
    -output="bacon_{{.OS}}_{{.Arch}}" \
    github.com/troykinsella/bacon
}

# Is a tag build?
if [ "$TRAVIS_PULL_REQUEST" == "false" ] && [ -n "$TRAVIS_TAG" ]; then
  cross_compile
fi
