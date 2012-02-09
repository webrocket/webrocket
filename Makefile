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

clean: clean-tools

install: tools install-tools install-man
	-@$(ECHO) "\n\033[1;32mCONGRATULATIONS COMRADE!\033[0;32m\nWebRocket has been built, tested and installed!\033[0m\n"

format:
	-@$(ECHO) "\n-- \033[0;35mFormatting\033[0m"
	@go fmt ./...

tools: clean $(BUILD_MAN)
	-@$(ECHO) "\n-- \033[0;35mResolving dependencies\033[0m"
	@go get ./...
	-@$(ECHO) "\n-- \033[0;35mBuilding library and tools\033[0m"
	-@$(ECHO) "pkg/webrocket"
	@go build ./pkg/webrocket
	-@$(ECHO) "pkg/kosmonaut"
	@go build ./pkg/kosmonaut
	-@$(ECHO) "cmd/webrocket-server"
	@go build ./cmd/webrocket-server
	-@$(ECHO) "cmd/webrocket-admin"
	@go build ./cmd/webrocket-admin
	-@$(ECHO) "\n-- \033[0;35mRunning tests\033[0m"
	@go test ./...

clean-tools:
	-@$(ECHO) "-- \033[0;35mCleaning up\033[0m"
	@go clean ./...
	-@rm -rf webrocket-admin
	-@rm -rf webrocket-server

install-tools:
	-@$(ECHO) "\n-- \033[0;36mInstalling tools\033[0m"
	-@$(ECHO) "webrocket-server"
	@go install ./cmd/webrocket-server
	-@$(ECHO) "webrocket-admin"
	@go install ./cmd/webrocket-admin

install-packages:
	-@$(ECHO) "\n-- \033[0;36mInstalling packages\033[0m"
	-@$(ECHO) "github.com/webrocket/webrocket/pkg/kosmonaut"
	@go install ./pkg/kosmonaut

man: clean-man
	@$(MAKE) -C docs

install-man:
	-@$(ECHO) "\n-- \033[0;36mInstalling documentation\033[0m"
	@$(MAKE) -C docs install

clean-man:
	-@$(MAKE) -C docs clean