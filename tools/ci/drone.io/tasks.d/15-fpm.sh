#!/bin/sh

echo ">>> Installing FPM..."
sudo apt-get install ruby-dev gcc
gem install fpm

exit 0
