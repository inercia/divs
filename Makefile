
LIBTUNTAP_URL = https://github.com/LaKabane/libtuntap.git

# full go-stype URLs for the commands we want to build
CMDS=github.com/inercia/divs/cmd/divs

#################################################################
# main

all: tuntap/Makefile
	@echo "Building libtuntap"
	@make -C tuntap all
	@echo "Building DiVS"
	@for C in $(CMDS) ; do echo "... building $$C " ; go build $$C ; done

tuntap/Makefile:
	cd tuntap && cmake -Wno-dev .

clean: tuntap/Makefile
	@echo "Cleaning libtuntap"
	@make -C tuntap clean
	@echo "Cleaning DiVS"
	@go clean

#################################################################
# dependencies

tuntap/CMakeLists.txt:
	@echo "Obtaining libtuntap from $(LIBTUNTAP_URL)"
	git remote add -f tuntap $(LIBTUNTAP_URL)
	git subtree pull --prefix tuntap tuntap master --squash

subtree-pull: tuntap/CMakeLists.txt
	@echo "Pulling from subtree tuntap"
	git subtree pull --prefix tuntap tuntap master --squash



