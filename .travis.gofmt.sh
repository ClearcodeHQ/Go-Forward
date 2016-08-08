#!/bin/bash

cd $TRAVIS_BUILD_DIR
result=$(gofmt -s -d .)
if [ "$result" != "" ]; then
	echo "$result" >&2
	exit 1
fi
