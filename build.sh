#!/bin/bash

GOROOT=~/src/go

if [ ! -d out ]; then
  mkdir -p out
fi

$GOROOT/bin/6g -o out/gorun.6 gorun.go \
&& $GOROOT/bin/6l -o out/gorun out/gorun.6
