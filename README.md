![Bacon](https://troykinsella.github.io/bacon/bacon.png)
---

[![Version](https://badge.fury.io/gh/troykinsella%2Fbacon.svg)](https://badge.fury.io/gh/troykinsella%2Fbacon)
[![License](https://img.shields.io/github/license/troykinsella/bacon.svg)](https://github.com/troykinsella/bacon/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/troykinsella/bacon.svg?branch=master)](https://travis-ci.org/troykinsella/bacon)

A tool to watch files for changes and continuously react by running commands.

## Contents

1. [Features](#features)
1. [Installation](#installation)
1. [Usage](#usage)
    1. [TL;DR](#tldr)
    1. [Program Commands](#program-commands)
    1. [Shell Commands](#shell-commands)
    1. [On-Success Commands](#on-success-commands)
    1. [On-Failure Commands](#on-failure-commands)
    1. [Command Arguments](#command-arguments)
    1. [Watch Files](#watch-files)
        1. [Includes](#includes)
        1. [Excludes](#excludes)
1. [Output](#output)
    1. [Command Summary Line](#command-summary-line)
        1. [Custom Summary Line](#custom-summary-line)
    1. [Status Notifications](#status-notifications)
1. [Troubleshooting](#troubleshooting)
1. [Road Map](#road-map)
1. [Similar Tools](#similar-tools)
1. [License](#license)

## Features

* Compatible with tooling for any technology (i.e. go, ruby, node.js, java, etc.); `bacon` simply runs shell commands
* Control files to watch using extended globs
* Command status summary line
* Command status system notifications

## Installation

Checkout [releases](https://github.com/troykinsella/bacon/releases) and download the appropriate binary for your system.
Put the binary in a convenient place, such as `/usr/local/bin/bacon`.

Or, run these commands to download and install:
```bash
VERSION=0.0.7
OS=darwin # or linux, or windows
curl -SL -o /usr/local/bin/bacon https://github.com/troykinsella/bacon/releases/download/v${VERSION}/bacon_${OS}_amd64
chmod +x /usr/local/bin/bacon
```

Or, for [Go lang](https://golang.org/doc/code.html) projects, from your `GOPATH`:
```bash
go get github.com/troykinsella/bacon
```

Lastly, test the installation:
```bash
bacon -h
```

## Usage

Running `bacon` will watch your files, run commands when they've changed. As
soon as you run `bacon`, it will immediately execute your commands, without
waiting for a watched file to change.

### TL;DR

Here.
```bash
# Watch files matching **/* in the CWD, except for files in **/.*,
# and run a script when any of them change.
bacon -c ./run-me.sh

# Watch Go lang project source files and run "go test" when any change.
bacon -w 'src/github.com/you/project/**/*.go' \
      -c 'go test github.com/you/project/...'

# Watch Node.js project source files and run mocha tests when any change.
bacon -w '**/*.js' \
      -w '**/*.json' \
      -e node_modules \
      -c 'mocha test/unit/*.js'

# Watch a mixed set of files by including some then excluding from those.
bacon -w '**/*.sh' \
      -e '**/third-party' \
      -c ./test.sh

# Run commands only when pass or fail.
bacon -c check-syntax.sh \
      -c find-bugs.sh \
      -p notfiy-tom-vogel-of-great-success.sh \
      -f notify-tom-vogel-of-great-failure.sh

# Use long options for readability.
bacon --watch '**/*.sh' \
      --exclude '**/third-party' \
      --cmd ./test.sh

# Print the effective list of files for the given inclusions and exclusions.
bacon list -w '**/*.rb' \
           -e '**/.git' \
           -e '**/naughty'

# Run the given commands once, as bacon would after a watched file change,
# then exit.
bacon run -c ./unit-tests.sh \
          -c ./int-tests.sh \
          -p ./celebrate.sh
```

### Program Commands

`bacon` provides several different ways to run it by passing (or omitting) a command name.
"Command" in this context is not to be confused with the shell commands that are given to
`bacon` to execute with the `-c` option (see [Shell Commands](#shell-commands)).

Commands:

Command     | Description
----------- | -----------
`<omitted>` | Watch a set of files, and run the given shell commands when they change.
`list`      | Print a list of files matched by the given inclusion and exclusion glob expressions, and exit.
`run`       | Run the given shell commands as `bacon` would when watching files, and exit.

Run `bacon -h` for comprehensive usage.

### Shell Commands

When files change, `bacon` runs the shell commands that you pass with the `-c, --cmd` option:
```bash
bacon -c "go test github.com/you/project/..." \
      -c "./find-bugs.sh"
```

Commands are executed in the order supplied, and against the same working directory
in which you ran `bacon`. When running several commands, `bacon` will only consider the
execution as "passing" if all commands exit with code `0`. The first command that
fails (exits with non-`0`), will abort the execution of subsequent commands, and
mark the entire execution as "failing".

### On-Success Commands

Commands supplied with the `-p, --pass` option are executed only when all `-c` commands pass.
```bash
bacon -c "go test github.com/you/project/..." \
      -p "./notify-the-pentagon.sh"
```
These "pass" commands do not influence the final pass/fail result.

### On-Failure Commands

Commands passed with the `-f, --fail` option are executed when a `-c` command fails.
```bash
bacon -c "go test github.com/you/project/..." \
      -f ./send-email-to-microsoft.sh
```
These "fail" commands do not influence the final pass/fail result.

### Command Arguments

All commands are interpolated to substitute `$1` with the absolute path
of the file that changed to trigger the command execution.

```bash
bash -c "go test github.com/you/project/..." \
     -p 'go fmt $1'
```

Here, `bacon` is running `go fmt` against the file that was just changed, if tests pass.
Note: Be sure to pass `$1` in single quotes so that your shell doesn't interpret it
prior to being passed into `bacon`.

When commands are executed not as a result of a file change, such as immediately after
running `bacon` or when using `bacon run`, `$1` is substituted with an empty string ("").

### Watch Files

Files can be watched for changes. "Change", specifically,
means: When a file is written to. Creation and deletion changes are ignored.

Files are selected for watching using extended glob syntax (having support for **).
See the [bmatcuk/doublestar](https://github.com/bmatcuk/doublestar) documentation for glob syntax.
`bacon` does not follow symlinks in resolving matches. Globs that do not start with `/` are
considered relative to the CWD.

A list of include globs and a list of exclude globs can be passed into `bacon` to tell it what to watch.
First, the list of includes is expanded, then the result is passed through the excludes list to arrive
at the effective list of files to watch.

Use the `bacon list` command to print the effective watch list, and exit.

#### Includes

Without telling `bacon` otherwise, it includes `**/*`, which translates into
"every file below the CWD". Supply one or more alternate include globs with the
`-w, --watch` option:
```bash
bacon -w "src/github.com/you/project/**" \
      -c "go test github.com/you/project/..."
```

#### Excludes

`bacon` excludes `**/.*` by default, which omits any `.*` (dot) file or directory.
Pass one ore more alternate exclude globs with the `-e, --exclude` option:

```bash
bacon -w "src/github.com/you/project/**" \
      -e "src/github.com/you/project/no-watchee/**" \
      -c "go test github.com/you/project/..."
```

If you supply an exclusion, be sure to also supply the overridden `**/.*` default,
if that's desirable.

```bash
bacon -e "exclude-me/**" \
      -e "**/.*" \
      -c ./test-my-stuff.sh
```

## Output

By default, `bacon` only prints summary lines, clearing the screen in between command executions
to hide clutter. But, if you pass it `-o, --show-output`, it will print all command output continuously.
Regardless of this option, if an execution fails, the output and error streams of the failing
command are printed to `bacon`'s standard error.

### Command Summary Line

Since it takes more than a single glance to figure out from the command output if 
commands have passed or failed, `bacon` prints an ansii-coloured summary line after
executions.

When commands start executing, `bacon` prints this:
```
[19:31:40] → Running
```

After commands complete successfully, a passing status looks like this:
```
[19:31:42] ✓ Passed
```

Or, if any command fails:
```
[19:37:13] ✗ Failed
```

#### Custom Summary Line

`bacon` uses [Go lang templating](https://golang.org/pkg/text/template/)
to format the summary line. The default template is:
```
[{{ .timeStamp }}] {{ .colorStart }}{{ .statusSymbol }} {{ .status }}{{ .colorEnd }}
```

Use the `--summary-format` option to customize the summary line.

The following variables are available to the template:

Name           | Description
-------------- | -----------
`changedDir`   | The absolute path to the directory containing the `changedFile`.
`changedFile`  | The absolute path to the file that was changed to trigger the command execution, if any.
`colorEnd`     | Terminates a started ansii color string.
`colorStart`   | An opening ansii color code string that reflects the passing status. i.e. Green (passed), yellow (running), or red (failed).
`duration`     | A string representing how long the command sequence took to execute.
`status`       | One of "Running", "Passed", or "Failed".
`statusSymbol` | One of "→" (for running), "✓" (for passed), or "✗" (for failed).
`timeStamp`    | A current timestamp in the format: "h:m:s".

### Status Notifications

Sometimes you don't want to watch a terminal to see `bacon` output, you just
want to know when things break, and when they're fixed. That's where
status system notifications come in. Notifications look like this:

![Commands Recovered](https://troykinsella.github.io/bacon/notify_recover.png)

In order to not spam you with notifications for every watched file change,
`bacon` will only notify you when:

* Commands pass or fail for the first time
* Commands were failing, but are now passing
* Commands were passing, but are now failing

If you don't want notifications, pass the `--no-notify` option.

## Troubleshooting

### My file changes aren't being noticed

Are your inclusion/exclusion globs correct? See what `bacon` is effectively watching
with the `bacon list` command.

Are you watching more files than your operating system can support? 
Adjust your include (`-w`), and/or exclude (`-e`) options as necessary to reduce the match count.

### System notifications aren't working

System notifications are supported by [0xAX/notificator](https://github.com/0xAX/notificator).
Refer to this documentation to see if notifications are supported on your operating system.

### My commands are executing endlessly

The commands you're running are potentially modifying files, causing and endless execution loop.
Stop it. In the future `bacon` will detect endless build loops.

### Typing out all these options sucks! wtf fml

Make a shell script!

## Road Map

* Detect endless build loops

## Similar Tools

If `bacon` doesn't suit your need, maybe the excellent [Tonkpils/snag](https://github.com/Tonkpils/snag) will.

## License

MIT © Troy Kinsella
