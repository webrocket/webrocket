UNAME=$(shell uname -s)
SOURCE_DIR=$(shell pwd)
SCRIPTS_DIR=$(SOURCE_DIR)/scripts
EXTRA_DEPS=

ifndef VERBOSE
MAKEFLAGS+=--no-print-directory
endif

ifeq ($(UNAME),Darwin)
ECHO=echo
else
ECHO=echo -e
endif

ASCIIDOC=asciidoc
CAT=cat

all: tools
	-@$(ECHO) "\n\033[1;32mCONGRATULATIONS COMRADE!\033[0;32m\nWebRocket has been built and tested!\033[0m\n"

check: tools format
	-@$(ECHO) ""

clean: clean-tools

install: tools install-tools install-man
	-@$(ECHO) "\n\033[1;32mCONGRATULATIONS COMRADE!\033[0;32m\nWebRocket has been built, tested and installed!\033[0m\n"

format:
	-@$(ECHO) "\n\033[0;35m%%% Formatting\033[0m"
	@go fmt ./...

tools: clean $(BUILD_MAN)
	-@$(ECHO) "\033[0;35m%%% Resolving dependencies\033[0m"
	@go get -v ./...
	-@$(ECHO) "\n\033[0;35m%%% Building libraries and tools\033[0m"
	-@$(ECHO) "engine"
	@go build ./engine
	-@$(ECHO) "kosmonaut"
	@go build ./kosmonaut
	-@$(ECHO) "cmd/webrocket-server"
	@go build ./cmd/webrocket-server
	-@$(ECHO) "cmd/webrocket-admin"
	@go build ./cmd/webrocket-admin
	-@$(ECHO) "\n\033[0;35m%%% Running tests\033[0m"
	@go test ./...

clean-tools:
	@go clean ./...
	-@rm -rf webrocket-admin
	-@rm -rf webrocket-server

install-tools:
	-@$(ECHO) "\n\033[0;36m%%% Installing tools\033[0m"
	-@$(ECHO) "webrocket-server"
	@go install ./cmd/webrocket-server
	-@$(ECHO) "webrocket-admin"
	@go install ./cmd/webrocket-admin

install-packages:
	-@$(ECHO) "\n\033[0;36m%%% Installing packages\033[0m"
	-@$(ECHO) "github.com/webrocket/webrocket/kosmonaut"
	@go install ./kosmonaut

man: clean-man
	@$(MAKE) -C docs

install-man:
	-@$(ECHO) "\033[0;36mInstalling documentation\033[0m"
	@$(MAKE) -C docs install

clean-man:
	-@$(MAKE) -C docs clean
