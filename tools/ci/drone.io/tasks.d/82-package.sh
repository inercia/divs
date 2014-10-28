#!/bin/bash

export GOROOT=/home/ubuntu/go
export PATH=$GOROOT/bin:/home/ubuntu/bin:/usr/local/bin:$PATH

echo ">>> Packaging..."
echo ">>> (GOROOT=$GOROOT)"
make package-deb

