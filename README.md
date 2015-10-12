character
=========

This is a tool for various manipulations on characters, as characters rather
than full strings, to show names, encodings and more.

Patches welcome


Building
--------

Assuming that this repository is checked out into
`src/github.com/philpennnock/character` relative to an entry in your `$GOPATH`
list of directories:

```console
$ go get github.com/hamfist/deppy && deppy get && make
```

That should work for most people; assumes GNU make.  In more detail:

1. Get/install dependencies:
  * If you want to roll the dice and gamble, just fetch the current
    dependencies, whatever version they might be:
    + `go get -d`
  * If you want to use the same versions I've built against, use a dependency
    version manager; I'm happy to include several as long as (1) they don't
    conflict and (2) they don't pull code of differing copyrights and licenses
    into this repository.  For now, I'm using `deppy`, which should be
    familiar to anyone who used the original form of `godep`:
    + `go get github.com/hamfist/deppy`
    + `deppy get`
2. Build; two options:
  * `go build` -- idiomatic Go, should always work
  * `make` or `gmake` -- using GNU Make, should do extra steps such as embed
    version identifiers

