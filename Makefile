
LIBTUNTAP_URL = https://github.com/LaKabane/libtuntap.git

# full go-stype URLs for the commands we want to build
CMDS_BASE     = github.com/inercia/divs/cmd
CMDS          = divs

# cflags and ldflags for cgo (we cannot define them in the tuntap.go)
CGO_CFLAGS    = -I`pwd`/tuntap
CGO_LDFLAGS   = -L`pwd`/tuntap/lib

#################################################################
# main

all: tuntap/Makefile
	@echo "Building libtuntap"
	@make -C tuntap tuntap-static
	@echo "Building DiVS"
	@for C in $(CMDS) ; do echo "... building $$C " ; go build -v $(CMDS_BASE)/$$C ; done

tuntap/Makefile:
	cd tuntap && cmake -Wno-dev .

clean: tuntap/Makefile
	@echo "Cleaning libtuntap"
	@make -C tuntap clean
	@echo "Cleaning DiVS"
	@go clean
	@rm -f $(CMDS)

#################################################################
# dependencies

tuntap/CMakeLists.txt:
	@echo "Obtaining libtuntap from $(LIBTUNTAP_URL)"
	git remote add -f tuntap $(LIBTUNTAP_URL)
	git subtree pull --prefix tuntap tuntap master --squash

subtree-pull: tuntap/CMakeLists.txt
	@echo "Pulling from subtree tuntap"
	git subtree pull --prefix tuntap tuntap master --squash



