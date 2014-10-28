#!/bin/bash

GTAR=go1.3.3.linux-amd64.tar.gz

echo ">>> Installing Go..."
cd $HOME
[ -f $GTAR ]    || wget -q http://golang.org/dl/$GTAR
[ -d $HOME/go ] || tar -xvpf $GTAR

export GOROOT=/home/ubuntu/go
export PATH=$GOROOT/bin:/home/ubuntu/bin:$PATH

exit 0
