# This is an optional-use makefile, targetted at gmake, for various
# tasks; basic installation should _always_ be `go build` compatible.
#
# If need to support non-GNU make too, use a Makefile.common file and move
# logic around as needed.  Although at this point I'm inclined to nuke the
# makefile entirely.

# Use CANONICAL_CHARACTER_REPO in environ to override where this is checked out
# You'll probably also need to bulk-edit the Go src.

ifdef CANONICAL_CHARACTER_REPO
REPO_PATH=	$(CANONICAL_CHARACTER_REPO)
else
REPO_PATH=	github.com/philpennock/character
endif

# Set this via the cmdline to change the tables backend
TABLES=		tabular

# http://blog.jgc.org/2011/07/gnu-make-recursive-wildcard-function.html
rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))
# mine:
rwildnovendor=$(filter-out vendor/%,$(call rwildcard,$1,$2))

SOURCES=	$(call rwildnovendor,,*.go)
TOP_SOURCE=	main.go
BINARIES=	character
CRUFT=		dependency-graph.png

# The go binary to use; you might override on the command-line to be 'gotip'
GO_CMD ?= go
GO_LDFLAGS:=

ifndef REPO_VERSION
REPO_VERSION := $(shell ./.version)
endif
GO_LDFLAGS+= -X $(REPO_PATH)/commands/version.VersionString=$(REPO_VERSION)

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
else ifeq ($(TABLES),tabular)
BUILD_TAGS+= tabular
ifneq "$(wildcard vendor/go.pennock.tech/tabular )" ""
TABULAR_DIR=vendor/go.pennock.tech/tabular
else
TABULAR_DIR=../../../go.pennock.tech/tabular
endif
GO_LDFLAGS+= -X go.pennock.tech/tabular.LinkerSpecifiedVersion=$(shell $(TABULAR_DIR)/.version)
endif

.PHONY : all install help devhelp short_help cleaninstall \
	dep dependsgraph vet lint \
	check-no-GOPATH
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
	@echo " 'dep': fetch dependencies at locked versions (go dep)"
	@echo " 'help': you're looking at it"
	@echo " 'clean': remove outputs in source dir"
	@echo " 'cleaninstall': try to remove installed locations"
	@echo " 'devhelp': see more targets, for maintainers"

devhelp:
	@echo " 'dependsgraph': view dependencies"
	@echo " 'vet': go vet"
	@echo " 'lint': golint"
	@echo " 'test': go test, unit-tests"

$(BINARIES): $(TOP_SOURCE) $(SOURCES) dep
ifeq ($(REPO_VERSION),)
	@echo "Missing a REPO_VERSION"
	@false
endif
	@echo "Building version $(REPO_VERSION) ..."
	$(GO_CMD) build -o $@ -tags "$(BUILD_TAGS)" -ldflags "$(GO_LDFLAGS)" -v $<

install: $(TOP_SOURCE) $(SOURCES)
ifeq ($(REPO_VERSION),)
	@echo "Missing a REPO_VERSION"
	@false
endif
	@echo "Installing version $(REPO_VERSION) ..."
	rm -f "$(BIN_DIR_TOP)/$(BINARIES)"
	$(GO_CMD) install -tags "$(BUILD_TAGS)" -ldflags "$(GO_LDFLAGS)" -v $(REPO_PATH)

dep:
	dep ensure

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

vendor: Gopkg.lock
	dep ensure

LICENSES_all.txt: LICENSE.txt Gopkg.lock vendor
	rm -f ./LICENSES_all.txt tmplicpart tmplic
	for DIR in $$(dep status -f '{{.ProjectRoot}}{{"\n"}}'); do ( cd "vendor/$$DIR"; for F in NOTICE* LICEN[SC]E* PATENTS; do test -s "$$F" || continue; echo "~~~ $$F - $$DIR ~~~"; cat "./$$F"; done; ) > tmplicpart ; test -s tmplicpart || continue; echo "~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~"; cat tmplicpart; echo; done > tmplic
	( echo "~~~ $(REPO_PATH) ~~~"; cat LICENSE.txt tmplic ; ) > ./LICENSES_all.txt
	@rm -f tmplicpart tmplic

check-no-GOPATH:
	@if test -n "$(GOPATH)"; then echo >&2 "make: GOPATH is set, can't use this target"; exit 1; fi

show-versions:
	date
	uname -a
	git version
	go version
	./.version
	dep status

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
