UNAME=$(shell uname -s)
SOURCE_DIR=$(shell pwd)
BUILD_DIR=$(SOURCE_DIR)/build
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

# Build manual pages if WITH_MAN specified.
ifeq ($(WITH_MAN),1)
EXTRA_DEPS+=man
endif

MSG=$(SCRIPTS_DIR)/msg.sh
ASCIIDOC=asciidoc
CAT=cat

all: clean pkg server admin $(EXTRA_DEPS)
	@rm -rf _build
	@$(MSG) "gathering things together in *./build*"
	mkdir -p $(BUILD_DIR)/bin
	cp webrocket-admin/webrocket-admin $(BUILD_DIR)/bin
	cp webrocket-server/webrocket-server $(BUILD_DIR)/bin
	mkdir -p $(BUILD_DIR)/share/man/man1
	cp docs/*.1 $(BUILD_DIR)/share/man/man1/ &>/dev/null || true
	@$(ECHO) "\n\033[1;32mCONGRATULATIONS COMARADE!\033[0;32m\nWebRocket has been built into \033[1;32m$(BUILD_DIR)/\033[0m\n"
	@$(CAT) $(SOURCE_DIR)/INTRO; $(ECHO) ""

clean: clean-pkg clean-server clean-admin clean-deps clean-man
	rm -rf build

clean-deps: clean-gouuid clean-persival clean-gostepper

check: clean check-pkg
	@$(ECHO) "\n\033[32mALL THE TESTS PASSED!\033[0m"

format: format-server format-admin format-pkg

papers:
	-$(ASCIIDOC) -d article -o INSTALL.html INSTALL
	-$(ASCIIDOC) -d article -o NEWS.html NEWS
	-$(ASCIIDOC) -d article -o CONTRIBUTE.html CONTRIBUTE
	-$(ASCIIDOC) -d article -o README.html README

install:
	@$(ECHO) Not implemented!

# ./docs
man:
	@$(MSG) "cd *./docs*"
	-@$(MAKE) -C docs
clean-man:
	-@$(MAKE) -C docs clean
install-man:
	-@$(MAKE) -C docs install

# ./webrocket
pkg: gouuid persival
	@$(MSG) "cd *./webrocket*"
	@$(MAKE) -C webrocket
	cp webrocket/_obj/*.a .
clean-pkg:
	@$(MAKE) -C webrocket clean
	rm -f *.a
check-pkg: pkg
	@$(MSG) "cd *./webrocket*"
	@$(MAKE) -C webrocket test
format-pkg:
	-@$(MAKE) -C webrocket format

# ./webrocket-server
server: pkg gostepper
	@$(MSG) "cd *./webrocket-server*"
	@$(MAKE) -C webrocket-server
clean-server:
	@$(MAKE) -C webrocket-server clean
format-server:
	-@$(MAKE) -C webrocket-server format

# ./webrocket-admin
admin: pkg server gostepper
	@$(MSG) "cd *./webrocket-admin*"
	@$(MAKE) -C webrocket-admin
clean-admin:
	@$(MAKE) -C webrocket-admin clean
format-admin:
	-@$(MAKE) -C webrocket-admin format

# ./deps/gouuid
gouuid:
	@$(MSG) "cd *./deps/gouuid*"	
	@$(MAKE) -C deps/gouuid
	cp deps/gouuid/_obj/github.com/nu7hatch/*.a .
clean-gouuid:
	$(MAKE) -C deps/gouuid clean

# ./deps/persival
persival:
	@$(MSG) "cd *./deps/persival*"
	@$(MAKE) -C deps/persival
	cp deps/persival/_obj/github.com/nu7hatch/*.a .
clean-persival:
	$(MAKE) -C deps/persival clean

# ./deps/gostepper
gostepper:
	@$(MSG) "cd *./deps/gostepper*"
	@$(MAKE) -C deps/gostepper
	cp deps/gostepper/_obj/github.com/nu7hatch/*.a .
clean-gostepper:
	$(MAKE) -C deps/gostepper clean