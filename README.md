# autotest

[![Version](https://badge.fury.io/gh/troykinsella%2Fautotest.svg)](https://badge.fury.io/gh/troykinsella%2Fautotest)
[![License](https://img.shields.io/github/license/troykinsella/autotest.svg)](https://github.com/troykinsella/autotest/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/troykinsella/autotest.svg?branch=master)](https://travis-ci.org/troykinsella/autotest)

A tool to watch a Go lang workspace for source file changes and automatically
run tests with `go test`.

`autotest` continuously prints `go test` output, as well as:

* An easily readable pass/fail summary line
* Generates a system notification when tests begin to fail, and when they've recovered

System notifications are supported by [0xAX/notificator](https://github.com/0xAX/notificator).
Refer to this documentation to see if notifications are supported on your operating system.

## Installation

From your `GOPATH`:
```bash
go get github.com/troykinsella/autotest
```

## Usage

`autotest` can be executed from any directory with the same effect,
since it operates against `$GOPATH`.

### Test Arguments

Run `autotest` by passing the arguments you'd normally pass to `go test` with the `-a` option:
```bash
autotest -a github.com/you/project/...
```

Since the `-a` value is passed verbatim to `go test`, you can include any supported options. For example:
```bash
autotest -a "-v -cover github.com/you/project/..."
```

### Watch Targets

Files and directories (all files within a directory) can be watched for changes. "Change", specifically,
means: when a file is written to; creation and deletion changes are ignored.

Paths are selected for watching using extended glob syntax (support for **). 
See the [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar) documentation for glob syntax.
`autotest` does not follow symlinks in resolving matches. Globs that do not start with `/` are
considered relative to `$GOPATH/src`.

A list of include globs and a list of exclude globs can be passed into `autotest` to tell it what to watch.
First, the list of includes is expanded, then the result is passed through the excludes list to arrive
at the effective list of matched files.

Use the `-d` (debug) option to print the effective watch list.

#### Includes 

By default, `autotest` includes `**/*.go`, which translates into "every go source file below $GOHOME/src".
Supply one or more alternate include globs with the `-w` (watch) option:
```bash
autotest -w "github.com/you/project/**" \
         -c github.com/you/project/...
```

#### Excludes

`autotest` excludes `**/.git/**` by default, which omits any `.git` directory.
Pass one ore more alternate exclude globs with the `-e` option:
```bash
autotest -w "github.com/you/project/**" \
         -e "github.com/you/project/no-watchee/**" \
         -c github.com/you/project/...
```

### Automatic Style Formatting (Experimental)

Pass the `-f` option to enable formatting changed `*.go` files with `go fmt` prior to running tests.
This feature is experimental, and may have buggy behaviour on some systems.

## Troubleshooting

### My file changes aren't being noticed

Are you watching more files than your operating system can support?
As the default include glob is `$GOPATH/src/**/*.go`, every Go
package you've `go get`'ed is being watched unless you trim it down.

Use the `-d` option to show effective file matches. Adjust your include (`-w`), 
and/or exclude (`-e`) options as necessary to reduce the match count.

## Road Map

* Tests
* Optionally throttle test runs down to once per a configurable time duration
* Run a custom command upon test success or failure

## License

MIT © Troy Kinsella
