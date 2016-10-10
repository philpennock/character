character
=========

[![Continuous Integration](https://secure.travis-ci.org/philpennock/character.svg?branch=master)](http://travis-ci.org/philpennock/character)

This is a tool for various manipulations on characters, as characters rather
than full strings, to show names, encodings and more.

The tool is structured as a top-level command, options, and sub-commands which
can do different things.  Many sub-commands will take a `-v` verbose option,
which gives more detail in a pretty-printed table.

```console
$ character help
[...]
$ character version
[...]
$ character name ✓
CHECK MARK
$ character named -h
[...]
$ character named 'CHECK MARK'
✓
$ character named -j CHECK MARK
✓
$ character named -v/ check
[... table of results of substring search ...]
$ character transform fraktur Hello world
ℌ𝔢𝔩𝔩𝔬 𝔴𝔬𝔯𝔩𝔡
```

[Licensed](./LICENSE.txt) under a MIT-style license.  
[Accumulated licenses of all dependencies](./LICENSES_all.txt) are available
too.  
Patches welcome.


Building
--------

This code should be `go get` compatible, if you don't care about versions of
library dependencies which have been confirmed to work.  If you encounter
problems, please try one of the vendored library mechanisms listed below
instead.

So `go get github.com/philpennock/character` should download and install the
tool into the `bin` dir inside the first directory listed in `$GOPATH`.


### GOPATH avoidance

*For non-Go-programmers*:
If you just `git clone` this repository and don't want to learn about the
`GOPATH` and don't want to set up a new global one, then run:

```console
$ make shuffle-and-build
```

*BEWARE* that this will move this directory, to be inside a child tree.
Also this uses the otherwise-optional `make` system, on the assumption that
this is okay for people who use Git but not Golang.  Note that GNU make
is required.

This target will:

1. Make a `go` directory as a sibling to the current directory, move the
   current directory into the correct place in the tree relative to that `go`
   dir and use that dir as the `GOPATH` for other tooling
2. Fetch and install `govendor` if not already available
3. Invoke `govendor` to sync the dependencies into the `vendor/` directory
4. Build `character`, leaving the binary in the now-current directory

Remember to `cd "$(pwd)"` afterwards.


### Dependencies and vendoring.

The dependent code is not checked into this repository, for various
philosophical reasons which I hold to be good.  If one of the dependencies
disappears, please file an Issue and I'll take care of managing the
dependency.

This is Golang, the binaries shouldn't be built live from source in
Production, it's acceptable for the disappearance of a project to make the
code briefly harder to compile.  Reproducible builds are only guaranteed for
official non-git releases, which _will_ include all dependencies.

Specifics available below, after the Building sub-section


### Building

Two simple options, in preference order:

1. Run `make` (or `gmake`) -- using GNU make, build with any extra steps such
   as embedding version identifiers in the resulting binary.
2. Run `go build` -- idiomatic Go, should always work.

You're done.


### go get

This is the "roll the dice and gamble" approach, which _usually_ works, but
not always.  Just fetch the current dependencies, whatever version they might
be, and build.

```console
$ go get -v
```

And then build.


### Deppy

Deppy is a simple vendor-locking tool for recording which versions are
known-good and making it easy to check those back out.  It's a fork from
`godep`, preserving its original working model.
Deppy plays well with Travis CI.

The `Deps` file used by `deppy` should be considered to be akin to a
`foo.lock` file in Ruby's ecosystem, but where the required files are not
listed in `foo` but instead taken straight from the imports of the Go code.

Assuming that this repository is checked out into
`src/github.com/philpennnock/character` relative to an entry in your `$GOPATH`
list of directories:

```console
$ go get github.com/hamfist/deppy
$ deppy restore
```

That should work for most people; assumes GNU make.  Without `make`, just
use `go build` and accept the loss of version information.


### govendor

govendor is a more sophisticated tool for managing locking and vendoring.
This is the tooling approach used for the `make shuffle-and-build` target.

```console
$ go get -v github.com/kardianos/govendor

  ... AND THEN ONE OF: ...

$ make gvsync
  OR
$ govendor sync +vendor +missing
```


### Another dependency version manager

I'm happy to supply dependency versioning lock-files for more tools, as
long as they don't conflict with any other tool, and they don't pull code
maintained by others into this repository.

I don't want code of multiple copyrights and licenses in one git repository.
I don't want someone else's misbehaviour coming to light to force me to
rewrite my git history to remove copyright violating code: that's worse for
provenance than having to hunt around for another clone of a dependency which
does at least match the available checksums.


Optional Tools
--------------

If you're just installing, you don't need these.  If you're developing, you
might.

* `go get -v github.com/kisielk/godepgraph`  
   Dependency graphing tool
* `go get -v github.com/golang/lint/golint`  
  Linter for Go
* `go get -v github.com/hamfist/deppy`  
  Dependency version manager; one choice
* `go get -v github.com/kardianos/govendor`  
  Dependency version manager; another choice

