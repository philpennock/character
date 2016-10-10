# This is an optional-use makefile, targetted at gmake, for various
# tasks; basic installation should _always_ be `go build` compatible.
#
# If need to support non-GNU make too, use a Makefile.common file and move
# logic around as needed.

# Use CANONICAL_CHARACTER_REPO in environ to override where this is checked out
# You'll probably also need to bulk-edit the Go src.
# The GOPATH-less shuffle logic will get upset at you if you fork to a new repo
# and don't set this environment variable.

ifdef CANONICAL_CHARACTER_REPO
REPO_PATH=	$(CANONICAL_CHARACTER_REPO)
else
REPO_PATH=	github.com/philpennock/character
endif

# Set this via the cmdline to change the tables backend
TABLES=		apcera

SOURCES=	$(shell find . -name vendor -prune -o -type f -name '*.go')
TOP_SOURCE=	main.go
BINARIES=	character
CRUFT=		dependency-graph.png

# The go binary to use; you might override on the command-line to be 'gotip'
GO_CMD ?= go

VERSION_VAR := $(REPO_PATH)/commands/version.VersionString
ifndef REPO_VERSION
REPO_VERSION := $(shell git describe --always --dirty --tags)
endif

# Where various files are installed
PKG_DIR_TOP := $(firstword $(subst :, ,$(GOPATH)))/pkg/$(shell go env GOOS)_$(shell go env GOARCH)
BIN_DIR_TOP := $(firstword $(subst :, ,$(GOPATH)))/bin

# Which platform are we on?
PLATFORM := $(shell uname -s)

# Collect the build-tags we want
BUILD_TAGS:=
ifeq ($(TABLES),apcera)
BUILD_TAGS+= termtables
else ifeq ($(TABLES),termtables)
BUILD_TAGS+= termtables
else ifeq ($(TABLES),tablewriter)
BUILD_TAGS+= tablewriter
endif

.PHONY : all install help devhelp short_help cleaninstall gvsync depends \
	dependsgraph vet lint \
	perform-shuffle check-no-GOPATH shuffle-and-build setgo-and-build
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
	@echo " 'install': install, currently just program, via go install"
	@echo " 'gvsync': fetch dependencies at locked versions (govendor)"
	@echo " 'depends': fetch dependencies at locked versions (deppy)"
	@echo " 'shuffle-and-build': build without a GOPATH setup"
	@echo " 'help': you're looking at it"
	@echo " 'clean': remove outputs in source dir"
	@echo " 'cleaninstall': try to remove installed locations"
	@echo " 'devhelp': see more targets, for maintainers"

devhelp:
	@echo " 'dependsgraph': view dependencies"
	@echo " 'vet': go vet"
	@echo " 'lint': golint"
	@echo " 'test': go test, unit-tests"
	@echo " 'gopathupdate': update dependencies in non-vendored GOPATH"
	@echo " 'depsync': sync various dependency files"

$(BINARIES): $(TOP_SOURCE) $(SOURCES)
ifeq ($(REPO_VERSION),)
	@echo "Missing a REPO_VERSION"
	@false
endif
	@echo "Building version $(REPO_VERSION) ..."
	$(GO_CMD) build -o $@ -tags "$(BUILD_TAGS)" -ldflags "-X $(VERSION_VAR)=$(REPO_VERSION)" -v $<

install: $(TOP_SOURCE) $(SOURCES)
ifeq ($(REPO_VERSION),)
	@echo "Missing a REPO_VERSION"
	@false
endif
	@echo "Installing version $(REPO_VERSION) ..."
	rm -f "$(BIN_DIR_TOP)/$(BINARIES)"
	$(GO_CMD) install -tags "$(BUILD_TAGS)" -ldflags "-X $(VERSION_VAR)=$(REPO_VERSION)" -v $(REPO_PATH)

gvsync:
	govendor sync +vendor +missing

depends:
	deppy restore

dependsgraph: dependency-graph.png
ifeq ($(PLATFORM),Darwin)
	open dependency-graph.png
else
	xdg-open dependency-graph.png
endif

dependency-graph.png:
	@echo If godepgraph is not installed: go get github.com/kisielk/godepgraph
	godepgraph -s . | dot -Tpng -o$@

list-depends-all-go:
	@go list -f '{{range .Deps}}{{printf "%s\n" .}}{{end}}' .

vet:
	@go vet ./...
	@echo done vet

lint:
	@echo "If golint is not installed: go get -v github.com/golang/lint/golint"
	@echo "We do not follow all style suggestions; in particular, consts are ALL_CAPS"
	@golint ./...
	@echo done lint

test:
	@go test ./...

versiontag:
	git tag -s -m "Version $(TAGVERSION)" "v$(TAGVERSION)"

clean:
	rm -fv $(BINARIES) $(CRUFT)

cleaninstall:
ifdef REPO_PATH
	rm -rfv "$(PKG_DIR_TOP)/$(REPO_PATH)" "$(BIN_DIR_TOP)/$(BINARIES)"
else
	@echo "MISSING REPO_PATH DEFINITION"
	@false
endif

# govendor has nicer tooling, let's treat that as authoritative

depsync: LICENSES_all.txt Deps
	@true

LICENSES_all.txt: LICENSE.txt vendor/vendor.json
	@# `govendor license` picks up the empty (freshly-truncated) file. If
	@# not truncated, would recurse. So just nuke the file before generation
	@# and ensure during-generation the filename doesn't match license-based
	@# naming
	rm -f ./LICENSES_all.txt
	govendor license > ./tmplic
	mv ./tmplic ./LICENSES_all.txt

Deps: vendor/vendor.json
	mv vendor vendor.-
	deppy save
	mv vendor.- vendor

vendor/vendor.json: $(SOURCES)
	govendor update +vendor

gopathupdate:
	mv vendor vendor.-
	go get -d -u -v
	mv vendor.- vendor

check-no-GOPATH:
	@if test -n "$(GOPATH)"; then echo >&2 "make: GOPATH is set, can't use this target"; exit 1; fi

perform-shuffle: check-no-GOPATH
	sh -x ./.shuffle-gopath

setgo-and-build:
	sh ./.shuffle-env-run make gvsync all

shuffle-and-build: perform-shuffle setgo-and-build

# Where BSD lets you `make -V VARNAME` to print the value of a variable instead
# of building a target, this gives GNU make a target `print-VARNAME` to print
# the value.  I have so missed this when using GNU make.
#
# This rule comes from a comment on
#   <http://blog.jgc.org/2015/04/the-one-line-you-should-add-to-every.html>
# where the commenter provided the shell meta-character-safe version.
print-%: ; @echo '$(subst ','\'',$*=$($*))'


# NOTE WELL:
# When I move to making tarball releases available, remember that I have
# committed in the README to copying a vendored source tree into the tarballs,
# for purposes of reproducible builds.
# Should probably also hard-code the version number in a generated .go file
# as part of that flow.
