# This is an optional-use makefile, targetted at gmake, for various
# tasks; basic installation should _always_ be `go build` compatible.
#
# If need to support non-GNU make too, use a Makefile.common file and move
# logic around as needed.

REPO_PATH=	github.com/philpennock/character

SOURCES=	$(shell find . -type f -name '*.go')
TOP_SOURCE=	main.go
BINARIES=	character
CRUFT=		dependency-graph.png

# The go binary to use; you might override on the command-line to be 'gotip'
GO_CMD ?= go

VERSION_VAR := github.com/philpennock/character/commands/version.VersionString
ifndef REPO_VERSION
REPO_VERSION := $(shell git describe --always --dirty --tags)
endif

# Where various files are installed
PKG_DIR_TOP := $(firstword $(subst :, ,$(GOPATH)))/pkg/$(shell go env GOOS)_$(shell go env GOARCH)
BIN_DIR_TOP := $(firstword $(subst :, ,$(GOPATH)))/bin

# Which platform are we on?
PLATFORM := $(shell uname -s)

.PHONY : all install help devhelp short_help cleaninstall depends dependsgraph \
	vet lint
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
	@echo " 'depends': fetch dependencies at locked versions"
	@echo " 'help': you're looking at it"
	@echo " 'clean': remove outputs in source dir"
	@echo " 'cleaninstall': try to remove installed locations"
	@echo " 'devhelp': see more targets, for maintainers"

devhelp:
	@echo " 'dependsgraph': view dependencies"
	@echo " 'vet': go vet"
	@echo " 'lint': golint"
	@echo " 'test': go test, unit-tests"

$(BINARIES): $(TOP_SOURCE) $(SOURCES)
ifeq ($(REPO_VERSION),)
	@echo "Missing a REPO_VERSION"
	@false
endif
	@echo "Building version $(REPO_VERSION) ..."
	$(GO_CMD) build -o $@ -ldflags "-X $(VERSION_VAR)=$(REPO_VERSION)" -v $<

install: $(TOP_SOURCE) $(SOURCES)
ifeq ($(REPO_VERSION),)
	@echo "Missing a REPO_VERSION"
	@false
endif
	@echo "Installing version $(REPO_VERSION) ..."
	rm -f "$(BIN_DIR_TOP)/$(BINARIES)"
	$(GO_CMD) install -ldflags "-X $(VERSION_VAR)=$(REPO_VERSION)" -v $(REPO_PATH)

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

# Where BSD lets you `make -V VARNAME` to print the value of a variable instead
# of building a target, this gives GNU make a target `print-VARNAME` to print
# the value.  I have so missed this when using GNU make.
#
# This rule comes from a comment on
#   <http://blog.jgc.org/2015/04/the-one-line-you-should-add-to-every.html>
# where the commenter provided the shell meta-character-safe version.
print-%: ; @echo '$(subst ','\'',$*=$($*))'
