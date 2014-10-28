#!/bin/sh

echo ">>> Installing FPM..."
sudo apt-get install ruby ruby-dev gcc
[ $? -eq 0 ] || (echo "ERROR: could not install packages" ; exit 1; )

sudo gem install fpm
[ $? -eq 0 ] ||  || (echo "ERROR: could not install FPM" ; exit 1; )

exit 0
