package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

const (
	defaultShell = "bash"
)

type E struct {
	commands     []string
	passCommands []string
	failCommands []string

	shell      string
	showOutput bool

	out io.Writer
	err io.Writer

	mu      *sync.Mutex
	first   bool
	passing bool
}

type Result struct {
	Passing    bool
	WasPassing bool
	First      bool
}

func New(
	commands []string,
	passCommands []string,
	failCommands []string,
	shell string,
	showOutput bool) *E {

	if shell == "" {
		shell = defaultShell
	}

	return &E{
		commands:     commands,
		passCommands: passCommands,
		failCommands: failCommands,

		shell:      shell,
		showOutput: showOutput,

		out: os.Stdout,
		err: os.Stderr,

		mu:      &sync.Mutex{},
		first:   true,
		passing: true,
	}
}

func (e *E) RunCommands(args []string) *Result {
	pass := true

	for _, cmd := range e.commands {
		err := e.runCommand(cmd, args)
		if err != nil {
			pass = false
			break
		}
	}

	if pass {
		e.runPassCommands(args)
	} else {
		e.runFailCommands(args)
	}

	e.mu.Lock()
	first := e.first
	e.first = false
	wasPassing := e.passing
	e.passing = pass
	e.mu.Unlock()

	return &Result{
		Passing:    pass,
		WasPassing: wasPassing,
		First:      first,
	}
}

func (e *E) runPassCommands(args []string) {
	for _, cmd := range e.passCommands {
		e.runCommand(cmd, args)
	}
}

func (e *E) runFailCommands(args []string) {
	for _, cmd := range e.failCommands {
		e.runCommand(cmd, args)
	}
}

func (e *E) makeCommand(cmd string, args []string) *exec.Cmd {
	argLen := int64(len(args))
	if args != nil && argLen > 0 {
		cmd = os.Expand(cmd, func(token string) string {
			i, err := strconv.ParseInt(token, 10, 0)
			if err == nil && i < argLen {
				return args[i]
			}
			return ""
		})
	}

	return exec.Command(e.shell, "-c", cmd)
}

func (e *E) runCommand(str string, args []string) error {

	args = append([]string{str}, args...)
	cmd := e.makeCommand(str, args)

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	if e.showOutput {
		cmd.Stdout = e.out
	} else {
		cmd.Stdout = &outBuf
	}
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if !e.showOutput && err != nil {
		fmt.Fprintf(e.err, "%s", &outBuf)
	}
	fmt.Fprintf(e.err, "%s", &errBuf)

	return err
}
