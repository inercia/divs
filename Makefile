
# current version
VERSION=0.1.0

# Protocol buffers args
PB_PROTO   := $(wildcard server/protobuf/*.proto)
PB_GO      := $(patsubst %.proto,%.pb.go,$(PB_PROTO))
PB_GO_TEST := $(patsubst %.proto,%pb_test.go,$(PB_PROTO))

PROTOC_ARGS = --proto_path=${GOPATH}/src \
			  --proto_path=${GOPATH}/src/code.google.com/p/gogoprotobuf/protobuf \
			  --proto_path=.

#################################################################
# main

all: divs

divs: $(PB_GO) FORCE
	@echo "Building DiVS"
	go build github.com/inercia/divs/cmd/divs

test: divs
	go test ./...

clean:
	@echo "Cleaning DiVS"
	@go clean
	rm -f divs $(PB_GO) $(PB_GO_TEST)
	rm -f divs*.pkg divs*.deb
	rm -f *~ */*~

${GOPATH}/bin/protoc-gen-gogo:
	@echo "Installing $$GOPATH/bin/protoc-gen-gogo"
	go get code.google.com/p/gogoprotobuf/proto
	go get code.google.com/p/gogoprotobuf/protoc-gen-gogo
	go get code.google.com/p/gogoprotobuf/gogoproto

%.pb.go %pb_test.go : %.proto  ${GOPATH}/bin/protoc-gen-gogo
	@echo "Generating code for Protocol Buffers definition: $<"
	PATH=${GOPATH}/bin:${PATH} protoc $(PROTOC_ARGS) --gogo_out=. $<

#################################################################
# deps

get: deps
deps:
	@echo "Getting all dependencies..."
	go get -d ./...

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
	divs=/usr/local/bin/divs \
	conf/divs.conf=/usr/local/etc/divs/divs.conf

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
