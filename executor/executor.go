package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
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

func (e *E) RunCommands(changed string, args []string) *Result {
	pass := true

	for _, cmd := range e.commands {
		err := e.runCommand(changed, cmd, args)
		if err != nil {
			pass = false
			break
		}
	}

	if pass {
		e.runPassCommands(changed, args)
	} else {
		e.runFailCommands(changed, args)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	first := e.first
	e.first = false
	wasPassing := e.passing
	e.passing = pass

	return &Result{
		Passing:    pass,
		WasPassing: wasPassing,
		First:      first,
	}
}

func (e *E) runPassCommands(changed string, args []string) {
	for _, cmd := range e.passCommands {
		e.runCommand(changed, cmd, args)
	}
}

func (e *E) runFailCommands(changed string, args []string) {
	for _, cmd := range e.failCommands {
		e.runCommand(changed, cmd, args)
	}
}

func (e *E) makeCommand(cmdStr string, args []string) *exec.Cmd {
	cmd := exec.Command(e.shell, "-c", cmdStr)
	cmd.Env = os.Environ()
	return cmd
}

func (e *E) runCommand(changed string, str string, args []string) error {
	args = append([]string{str}, args...)
	cmd := e.makeCommand(str, args)
	if changed != "" {
		cmd.Env = append(cmd.Env, "BACON_CHANGED=" + changed)
	}

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
