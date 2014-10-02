
# full go-stype URLs for the commands we want to build
CMDS_BASE     = github.com/inercia/divs/cmd
CMDS          = divs

#################################################################
# main

all:
	@echo "Building DiVS"
	@for C in $(CMDS) ; do echo "... building $$C " ; \
	    CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" \
	        go build $(CMDS_BASE)/$$C ; \
	done

clean:
	@echo "Cleaning DiVS"
	@go clean
	@rm -f $(CMDS)

