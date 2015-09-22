# This is an optional-use makefile, targetted at gmake, for various
# tasks; basic installation should _always_ be `go build` compatible.
#
# If need to support non-GNU make too, use a Makefile.common file and move
# logic around as needed.

REPO_PATH=	github.com/philpennock/character

SOURCES=	main.go
BINARIES=	character

# The go binary to use; you might override on the command-line to be 'gotip'
GO_CMD ?= go

VERSION_VAR := github.com/philpennock/character/commands/version.VersionString
ifndef REPO_VERSION
REPO_VERSION := $(shell git describe --always --dirty --tags)
endif

# Where various files are installed
PKG_DIR_TOP := $(firstword $(subst :, ,$(GOPATH)))/pkg/$(shell go env GOOS)_$(shell go env GOARCH)
BIN_DIR_TOP := $(firstword $(subst :, ,$(GOPATH)))/bin

.PHONY : all help short_help cleaninstall
.DEFAULT_GOAL := helpful_all

# first build target references hint to extra help
helpful_all: short_help all

all: $(BINARIES)

short_help:
	@echo "*** You can try 'make help' for hints on targets ***"
	@echo

help:
	@echo "The following targets are available:"
	@echo " 'all': make all programs"
	@echo " 'help': you're looking at it"
	@echo " 'clean': remove outputs in source dir"
	@echo " 'cleaninstall': try to remove installed locations"

$(BINARIES): $(SOURCES)
ifeq ($(REPO_VERSION),)
	@echo "Missing a REPO_VERSION"
	@false
endif
	@echo "Building version $(REPO_VERSION) ..."
	$(GO_CMD) build -o $@ -ldflags "-X $(VERSION_VAR)=$(REPO_VERSION)" -v $<

clean:
	rm -fv $(BINARIES)

cleaninstall:
ifdef REPO_PATH
	rm -rfv "$(PKG_DIR_TOP)/$(REPO_PATH)" "$(BIN_DIR_TOP)/$(BINARIES)"
else
	@echo "MISSING REPO_PATH DEFINITION"
	@false
endif

