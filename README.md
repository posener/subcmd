# subcmd

[![Build Status](https://travis-ci.org/posener/subcmd.svg?branch=master)](https://travis-ci.org/posener/subcmd)
[![codecov](https://codecov.io/gh/posener/subcmd/branch/master/graph/badge.svg)](https://codecov.io/gh/posener/subcmd)
[![GoDoc](https://godoc.org/github.com/posener/subcmd?status.svg)](http://godoc.org/github.com/posener/subcmd)
[![goreadme](https://goreadme.herokuapp.com/badge/posener/subcmd.svg)](https://goreadme.herokuapp.com)

subcmd is a minimalistic library that enables easy sub commands with the standard `flag` library.

Define a `root` command object using the `Root` function.
This object exposes the standard library's `flag.FlagSet` API, which enables adding flags in the
standard way.
Additionally, this object exposes the `SubCommand` method, which returns another command object.
This objects also exposing the same API, enabling definition of flags and nested sub commands.

The root object then have to be called with the `Parse` or `ParseArgs` methods, similiraly to
the `flag.Parse` call.

The usage is automatically configured to show both sub commands and flags.

#### Positional arguments

The `subcmd` library is opinionated about positional arguments: it enforces their definition
and parsing. The user can define for each sub command if and how many positional arguments it
accepts. Their usage is similar to the flag values usage.

#### Example

See [./example/main.go](./example/main.go).

#### Limitations

Suppose `cmd` has a flag `-flag`, and a subcommand `sub`. In the current implementation:
Calling `cmd sub -flag` won't work as the flag is set after the sub command, while
`cmd -flag sub` will work perfectly fine. Each flag needs to be used in the scope of its command.

## Functions

### func [OptDetails](https://github.com/posener/subcmd/blob/master/subcmd.go#L137)

`func OptDetails(details string) optionFn`

OptSynopsis sets a description to the root command.

### func [OptErrorHandling](https://github.com/posener/subcmd/blob/master/subcmd.go#L109)

`func OptErrorHandling(errorHandling flag.ErrorHandling) optionRootFn`

OptErrorHandling defines the behavior in case of an error in the `Parse` function.

### func [OptName](https://github.com/posener/subcmd/blob/master/subcmd.go#L123)

`func OptName(name string) optionRootFn`

OptName sets a predefined name to the root command.

### func [OptOutput](https://github.com/posener/subcmd/blob/master/subcmd.go#L116)

`func OptOutput(w io.Writer) optionRootFn`

OptOutput sets the output for the usage.

### func [OptSynopsis](https://github.com/posener/subcmd/blob/master/subcmd.go#L130)

`func OptSynopsis(synopsis string) optionRootFn`

OptSynopsis sets a description to the root command.

## Sub Packages

* [example](./example)


---

Created by [goreadme](https://github.com/apps/goreadme)
