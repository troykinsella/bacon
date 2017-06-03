package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	defaultShell = "bash"
)

type E struct {
	commands     []string
	passCommands []string
	failCommands []string

	shell      string
	dir        string
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
	Duration   time.Duration
	FinishedAt time.Time
}

func New(
	commands []string,
	passCommands []string,
	failCommands []string,
	shell string,
	dir string,
	showOutput bool) *E {

	if shell == "" {
		shell = defaultShell
	}

	return &E{
		commands:     commands,
		passCommands: passCommands,
		failCommands: failCommands,

		shell:      shell,
		dir:        dir,
		showOutput: showOutput,

		out: os.Stdout,
		err: os.Stderr,

		mu:      &sync.Mutex{},
		first:   true,
		passing: true,
	}
}

func (e *E) RunCommands(
	changed string,
	args []string) *Result {

	start := time.Now()
	pass := true

	for _, cmd := range e.commands {
		err := e.runCommand(changed, cmd, args)
		if err != nil {
			pass = false
			break
		}
	}

	passFailCommands := e.passCommands
	if !pass {
		passFailCommands = e.failCommands
	}
	for _, cmd := range passFailCommands {
		e.runCommand(changed, cmd, args)
	}

	end := time.Now()
	duration := end.Sub(start)

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
		Duration:   duration,
		FinishedAt: end,
	}
}

func (e *E) makeCommand(changed string, cmdStr string, args []string) *exec.Cmd {
	cmd := exec.Command(e.shell, "-c", cmdStr)

	cmd.Env = os.Environ()
	if changed != "" {
		cmd.Env = append(cmd.Env, "BACON_CHANGED=" + changed)
	}

	if e.dir != "" {
		cmd.Dir = e.dir
	}

	return cmd
}

func (e *E) runCommand(
	changed string,
	str string,
	args []string,
) error {
	args = append([]string{str}, args...)
	cmd := e.makeCommand(changed, str, args)

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
