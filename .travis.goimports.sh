#!/bin/bash

cd $TRAVIS_BUILD_DIR
result=$(goimports -d *.go)
if [ "$result" != "" ]; then
	echo "$result" >&2
	exit 1
fi
