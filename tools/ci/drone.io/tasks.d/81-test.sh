#!/bin/bash

export GOROOT=/home/ubuntu/go
export PATH=$GOROOT/bin:/home/ubuntu/bin:$PATH

echo ">>> Testing..."
echo ">>> (GOROOT=$GOROOT)"
make test

