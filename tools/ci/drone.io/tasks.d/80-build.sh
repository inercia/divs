#!/bin/bash

export GOROOT=/home/ubuntu/go
export PATH=$GOROOT/bin:/home/ubuntu/bin:$PATH

echo ">>> Building in" `pwd`
echo ">>> (GOROOT=$GOROOT)"
make deps
make all

