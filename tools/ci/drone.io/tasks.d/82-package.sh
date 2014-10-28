#!/bin/bash

export GOROOT=/home/ubuntu/go
export PATH=$GOROOT/bin:/home/ubuntu/bin:$PATH

echo ">>> Packaging..."
echo ">>> (GOROOT=$GOROOT)"
make package-deb

