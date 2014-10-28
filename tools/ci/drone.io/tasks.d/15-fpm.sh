#!/bin/sh

echo ">>> Installing FPM..."
sudo apt-get update

sudo apt-get install ruby1.9.1 ruby1.9.1-dev \
  rubygems1.9.1 irb1.9.1 ri1.9.1 rdoc1.9.1 \
  build-essential libopenssl-ruby1.9.1 libssl-dev zlib1g-dev
[ $? -eq 0 ] || (echo "ERROR: could not install packages" ; exit 1 ; )

sudo update-alternatives --install /usr/bin/ruby ruby /usr/bin/ruby1.9.1 400 \
         --slave   /usr/share/man/man1/ruby.1.gz ruby.1.gz \
                        /usr/share/man/man1/ruby1.9.1.1.gz \
        --slave   /usr/bin/ri ri /usr/bin/ri1.9.1 \
        --slave   /usr/bin/irb irb /usr/bin/irb1.9.1 \
        --slave   /usr/bin/rdoc rdoc /usr/bin/rdoc1.9.1
[ $? -eq 0 ] || (echo "ERROR: could not update alternatives" ; exit 1 ; )

# choose your interpreter
# changes symlinks for /usr/bin/ruby , /usr/bin/gem
# /usr/bin/irb, /usr/bin/ri and man (1) ruby
sudo update-alternatives --auto ruby
[ $? -eq 0 ] || (echo "ERROR: could not update alternatives" ; exit 1 ; )

sudo update-alternatives --auto gem
[ $? -eq 0 ] || (echo "ERROR: could not update alternatives" ; exit 1 ; )

sudo gem install fpm
[ $? -eq 0 ] || (echo "ERROR: could not install FPM" ; exit 1 ; )

exit 0
