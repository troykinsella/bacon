# autotest

[![Version](https://badge.fury.io/gh/troykinsella%2Fautotest.svg)](https://badge.fury.io/gh/troykinsella%2Fautotest)
[![License](https://img.shields.io/github/license/troykinsella/autotest.svg)](https://github.com/troykinsella/autotest/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/troykinsella/autotest.svg?branch=master)](https://travis-ci.org/troykinsella/autotest)

A tool to watch files for changes and react by running commands. 

## Installation

From your `GOPATH`:
```bash
go get github.com/troykinsella/autotest
```

Or, checkout [releases](https://github.com/troykinsella/autotest/releases) and download the appropriate binary for your system.

Or, run these commands:
```bash
VERSION=0.0.1
OS=linux # or darwin, or windows
curl -SsL -o $GOPATH/bin/autotest https://github.com/troykinsella/autotest/releases/download/v${VERSION}/autotest_${OS}_amd64
chmod +x $GOPATH/bin/autotest
```

## Usage

Running `autotest` will watch your files, run commands when they've changed.

Run `autotest -h` for comprehensive command usage.

### Commands

`autotest` runs the commands that you pass with the `-c` option:
```bash
autotest -c "go test github.com/you/project/..." -c "echo holy shit that's wicked"
```

### Passing Commands

Commands supplied with the `-p` option are executed only when all commands pass.
```bash
autotest -c "go test github.com/you/project/..." -p "echo haw yeah"
```

### Failing Commands

Commands passed with the `-f` option are executed when a command fails.
```bash
autotest -c "go test github.com/you/project/..." -f ./sendEmailToMicrosoft.sh
```

### Output

`autotest` continuously prints command output, as well as:

* An easily readable pass/fail summary line
* Generates a system notification when commands begin to fail, and when they've recovered

#### Command Summary Line

Since it takes more than a single glance to figure out from the command output if 
they've passed, `autotest` prints an ansii-coloured summary line after command executions.

Pass looks like this:
```
[19:31:42] ✓ Passed
```
The timestamp format is `[h:m:s]`.

Failures look like this:
```
[19:37:13] ✗ Failed
```

#### Status Notifications

In order to not spam you with notifications for every watched file change,
`autotest` will only notify you when:

* Commands pass or fail for the first time
* Commands were failing, but are now passing
* Commands were passing, but are now failing

Notifications look like this:

![Commands Recovered](https://troykinsella.github.io/autotest/notify_recover.png)

System notifications are supported by [0xAX/notificator](https://github.com/0xAX/notificator).
Refer to this documentation to see if notifications are supported on your operating system.
Notifications are always enabled in `autotest`, but if they fail, it's silent.

### Watch Targets

Files and directories (all files within a directory) can be watched for changes. "Change", specifically,
means: when a file is written to; creation and deletion changes are ignored.

Paths are selected for watching using extended glob syntax (support for **). 
See the [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar) documentation for glob syntax.
`autotest` does not follow symlinks in resolving matches. Globs that do not start with `/` are
considered relative to the CWD.

A list of include globs and a list of exclude globs can be passed into `autotest` to tell it what to watch.
First, the list of includes is expanded, then the result is passed through the excludes list to arrive
at the effective list of matched files.

Use the `-d` (debug) option to print the effective watch list.

#### Includes 

By default, `autotest` includes `**/*`, which translates into "every below the CWD".
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

## Troubleshooting

### My file changes aren't being noticed

Are you watching more files than your operating system can support? Use 
the `-d` option to show effective file matches. Adjust your include (`-w`), 
and/or exclude (`-e`) options as necessary to reduce the match count.

## Road Map

* Tests
* Optionally throttle test runs down to once per a configurable time duration

## Comparison to Similar Tools

If `autotest` doesn't suit your need, maybe the excellent [Tonkpils/snag](https://github.com/Tonkpils/snag) will.

## License

MIT © Troy Kinsella
