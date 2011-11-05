#!/bin/bash

GOROOT=~/src/go

if [ ! -d bin ]; then
  mkdir -p bin
fi

$GOROOT/bin/6g -o bin/gorun.6 gorun.go \
&& $GOROOT/bin/6l -o bin/gorun bin/gorun.6
