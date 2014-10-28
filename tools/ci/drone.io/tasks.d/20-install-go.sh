#!/bin/bash

REQUIRED_VERSION="go1.3.3"
INSTALLED_VERSION=$(go version | awk '{ print $3 }')

if [ "$REQUIRED_VERSION" != "$INSTALLED_VERSION" ] ; then
	GTAR=$REQUIRED_VERSION.linux-amd64.tar.gz

	echo ">>> Installing Go version $REQUIRED_VERSION..."
	pushd `pwd`
	cd $HOME
	[ -f $GTAR ]    || wget -q http://golang.org/dl/$GTAR
	[ -d $GOROOT  ] || tar -xvpf $GTAR
	popd
fi

export GOROOT=/home/ubuntu/go
export PATH=$GOROOT/bin:/home/ubuntu/bin:$PATH

exit 0
