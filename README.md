# bacon

[![Version](https://badge.fury.io/gh/troykinsella%2Fbacon.svg)](https://badge.fury.io/gh/troykinsella%2Fbacon)
[![License](https://img.shields.io/github/license/troykinsella/bacon.svg)](https://github.com/troykinsella/bacon/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/troykinsella/bacon.svg?branch=master)](https://travis-ci.org/troykinsella/bacon)

A tool to watch files for changes and react by running commands. 

## Installation

From your `GOPATH`:
```bash
go get github.com/troykinsella/bacon
```

Or, checkout [releases](https://github.com/troykinsella/bacon/releases) and download the appropriate binary for your system.

Or, run these commands:
```bash
VERSION=0.0.2
OS=linux # or darwin, or windows
curl -SsL -o /usr/local/bin/bacon https://github.com/troykinsella/bacon/releases/download/v${VERSION}/bacon_${OS}_amd64
chmod +x /usr/local/bin/bacon
```

## Usage

Running `bacon` will watch your files, run commands when they've changed.

Run `bacon -h` for comprehensive command usage.

### Commands

`bacon` runs the commands that you pass with the `-c` option:
```bash
bacon -c "go test github.com/you/project/..." -c "echo holy shit that's wicked"
```

### Passing Commands

Commands supplied with the `-p` option are executed only when all commands pass.
```bash
bacon -c "go test github.com/you/project/..." -p "echo haw yeah"
```

### Failing Commands

Commands passed with the `-f` option are executed when a command fails.
```bash
bacon -c "go test github.com/you/project/..." -f ./sendEmailToMicrosoft.sh
```

### Output

`bacon` continuously prints command output, as well as:

* An easily readable pass/fail summary line
* Generates a system notification when commands begin to fail, and when they've recovered

#### Command Summary Line

Since it takes more than a single glance to figure out from the command output if 
they've passed, `bacon` prints an ansii-coloured summary line after command executions.

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
`bacon` will only notify you when:

* Commands pass or fail for the first time
* Commands were failing, but are now passing
* Commands were passing, but are now failing

Notifications look like this:

![Commands Recovered](https://troykinsella.github.io/bacon/notify_recover.png)

System notifications are supported by [0xAX/notificator](https://github.com/0xAX/notificator).
Refer to this documentation to see if notifications are supported on your operating system.
Notifications are always enabled in `bacon`, but if they fail, it's silent.

### Watch Targets

Files and directories (all files within a directory) can be watched for changes. "Change", specifically,
means: when a file is written to; creation and deletion changes are ignored.

Paths are selected for watching using extended glob syntax (support for **). 
See the [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar) documentation for glob syntax.
`bacon` does not follow symlinks in resolving matches. Globs that do not start with `/` are
considered relative to the CWD.

A list of include globs and a list of exclude globs can be passed into `bacon` to tell it what to watch.
First, the list of includes is expanded, then the result is passed through the excludes list to arrive
at the effective list of matched files.

Use the `-d` (debug) option to print the effective watch list.

#### Includes 

By default, `bacon` includes `**/*`, which translates into "every below the CWD".
Supply one or more alternate include globs with the `-w` (watch) option:
```bash
bacon -w "github.com/you/project/**" \
      -c "go test github.com/you/project/..."
```

#### Excludes

`bacon` excludes `**/.git/**` by default, which omits any `.git` directory.
Pass one ore more alternate exclude globs with the `-e` option:
```bash
bacon -w "github.com/you/project/**" \
      -e "github.com/you/project/no-watchee/**" \
      -c "go test github.com/you/project/..."
```

## Troubleshooting

### My file changes aren't being noticed

Are you watching more files than your operating system can support? Use 
the `-d` option to show effective file matches. Adjust your include (`-w`), 
and/or exclude (`-e`) options as necessary to reduce the match count.

## Road Map

* Tests
* Optionally throttle command runs down to once per a configurable time duration

## Comparison to Similar Tools

If `bacon` doesn't suit your need, maybe the excellent [Tonkpils/snag](https://github.com/Tonkpils/snag) will.

## License

MIT © Troy Kinsella
