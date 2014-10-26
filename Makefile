
# current version
VERSION=0.1.0

# the go tool
GO=${GOROOT}/bin/go

#################################################################
# main

all: divsd.exe

divsd.exe: $(PB_GO) FORCE
	@echo "Building DiVS"
	$(GO) build -o divsd.exe github.com/inercia/divs/cmd/divsd

test: divsd.exe
	$(GO) test ./...

clean:
	@echo "Cleaning DiVS"
	@go clean
	rm -rf bin build
	rm -f divsd.exe $(PB_GO) $(PB_GO_TEST)
	rm -f divs*.pkg divs*.deb
	rm -f *~ */*~


#################################################################
# deps

get: deps
dependencies: deps
deps: clean
	@echo "Getting all dependencies..."
	$(GO) get -d ./...

dependencies-update: deps-up
deps-update: deps-up
deps-up: clean
	@echo "Updating all dependencies..."
	$(GO) get -d -u ./...

distclean-deps:
	for PKG in $$GOPATH/src/*/* ; do \
		if [ -d $$PKG ] ; then \
			[ `basename $$PKG` != "inercia" ] && rm -rf $$PKG ; \
		fi ; \
	done
	rm -rf $$GOPATH/pkg
	
#################################################################
# packaging

# in order to cross compile you must do this for
# each OS/architecture you want:
#
# $ cd $GOROOT/src
# $ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 ./make.bash --no-clean
#

PACKAGING_COMMON=\
	-s dir \
	-v $(VERSION) \
	-n divs \
	--config-files /usr/local/etc/divs/divs.conf \
	divsd.exe=/usr/local/bin/divsd \
	conf/etc/divsd.conf=/usr/local/etc/divs/divsd.conf

# install fpm with:
# $ gem install fpm
package: package-osx package-deb

package-osx:
	make clean
	GOOS=darwin GOARCH=amd64 make all
	fpm -t osxpkg $(PACKAGING_COMMON)

# on Mac: brew install gnu-tar
package-deb:
	make clean
	GOOS=linux GOARCH=amd64 make all
	fpm -t deb $(PACKAGING_COMMON)

#################################################################

FORCE:
