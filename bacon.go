package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/0xAX/notificator"
	"gopkg.in/fsnotify.v1"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type Options struct {
	commands     []string
	passCommands []string
	failCommands []string

	includes []string
	excludes []string

	showOutput bool
	notify     bool
	debug      bool
}

func NewOptions(commands []string,
	passCommands []string,
	failCommands []string,
	includes []string,
	excludes []string,
	showOutput bool,
    notify bool,
	debug bool) (*Options, error) {

	includes = normalizeGlobs(includes, "**/*")
	excludes = normalizeGlobs(excludes, "**/.git")

	if debug {
		fmt.Printf("Includes: %s\n", includes)
		fmt.Printf("Excludes: %s\n", excludes)
	}

	return &Options{
		commands:     commands,
		passCommands: passCommands,
		failCommands: failCommands,
		includes:     includes,
		excludes:     excludes,
		showOutput:   showOutput,
		notify:       notify,
		debug:        debug,
	}, nil
}

type AutoTest struct {
	watcher *fsnotify.Watcher
	opts    *Options
	n       *notificator.Notificator

	mu      *sync.Mutex
	first   bool
	passing bool
}

func NewAutoTest(o *Options) *AutoTest {
	return &AutoTest{
		opts: o,
		n:    newNotificator(),

		mu:      &sync.Mutex{},
		first:   true,
		passing: true,
	}
}

func newNotificator() *notificator.Notificator {
	return notificator.New(notificator.Options{
		AppName: "bacon",
	})
}

func (at *AutoTest) pathSelected(path string) bool {
	sel, err := selected(path, at.opts.includes, at.opts.excludes)
	if err != nil {
		return false
	}
	return sel
}

func (at *AutoTest) watchPaths(paths []string) error {
	for _, p := range paths {
		err := at.watchPath(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (at *AutoTest) watchPath(path string) error {
	err := at.watcher.Add(path)
	return err
}

func (at *AutoTest) unwatchPath(path string) error {
	err := at.watcher.Remove(path)
	return err
}

func extractName(e fsnotify.Event) string {
	n := e.Name

	// Lame-ass JetBrains hack: remove a "___jb_tmp___" suffix
	const sfx = "___jb_tmp___"
	if strings.HasSuffix(n, sfx) {
		n = n[0 : len(n)-len(sfx)]
	}

	// ___jb_old___ too?

	return n
}

func (at *AutoTest) newWatcher(done chan error) (*fsnotify.Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event := <-w.Events:
				n := extractName(event)
				if !at.pathSelected(n) {
					continue
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					//at.handleCreate(n)
				} else if event.Op&fsnotify.Write == fsnotify.Write {
					at.handleChange(n)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					//at.handleRemove(n)
				}
			case err := <-w.Errors:
				done <- err
			}
		}
	}()

	return w, nil
}

func (at *AutoTest) Run() error {
	done := make(chan error)
	w, err := at.newWatcher(done)
	if err != nil {
		return err
	}
	at.watcher = w

	defer w.Close()

	dirs, err := expandedBaseDirs(at.opts.includes, at.opts.excludes)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return errors.New("No paths to watch were matched")
	}
	if at.opts.debug {
		fmt.Println("Watching:")
		for _, t := range dirs {
			fmt.Println(t)
		}
	}

	at.watchPaths(dirs)

	if !at.opts.debug {
		at.runCommands() // don't wait for a change
	}

	return <-done
}

func (at *AutoTest) handleCreate(f string) {
	if at.opts.debug {
		fmt.Printf("Created: %s\n", f)
	}

	if sel := at.pathSelected(f); sel {
		if err := at.watchPath(f); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to watch new file: %s\n", f)
			return
		}

		if at.opts.debug {
			fmt.Printf("Watching: %s\n", f)
		}
	}
}

func (at *AutoTest) handleRemove(f string) {
	if at.opts.debug {
		fmt.Printf("Removed: %s\n", f)
	}

	at.unwatchPath(f)
}

func (at *AutoTest) handleChange(f string) {
	if at.opts.debug {
		fmt.Printf("Changed: %s\n", f)
	}

	// TODO: Throttle command executions here

	at.runCommands()
}

func (at *AutoTest) runCommands() {
	cls()
	printRunning()

	pass := true

	for _, cmd := range at.opts.commands {
		if at.opts.debug {
			fmt.Printf("Running: %s\n", cmd)
		}

		err := at.runCommand(cmd)
		if err != nil {
			pass = false
			break
		}
	}

	if pass {
		at.runPassCommands()
	} else {
		at.runFailCommands()
	}

	if !at.opts.showOutput && pass {
		cls()
	}
	printSummary(pass)

	at.mu.Lock()
	at.notify(pass)

	at.first = false
	at.passing = pass
	at.mu.Unlock()
}

func (at *AutoTest) runPassCommands() {
	for _, cmd := range at.opts.passCommands {
		if at.opts.debug {
			fmt.Printf("Pass: %s\n", cmd)
		}

		at.runCommand(cmd)
	}
}

func (at *AutoTest) runFailCommands() {
	for _, cmd := range at.opts.failCommands {
		if at.opts.debug {
			fmt.Printf("Fail: %s\n", cmd)
		}

		at.runCommand(cmd)
	}
}

func (at *AutoTest) runCommand(str string) error {


	cmd := exec.Command("sh", "-c", str)

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer

	if at.opts.showOutput {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = &outBuf
	}
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if !at.opts.showOutput && err != nil {
		cls()
		fmt.Fprintf(os.Stderr, "%s", &outBuf)
	}
	fmt.Fprintf(os.Stderr, "%s", &errBuf)

	return err
}

func (at *AutoTest) notify(result bool) {
	msg := at.notifyMessage(result)
	if msg != "" {
		at.n.Push("bacon", msg, "noop", notificator.UR_NORMAL)
	}
}

func (at *AutoTest) notifyMessage(result bool) string {
	if !at.opts.notify {
		return ""
	}

	var msg string
	if at.first {
		if result {
			msg = "✓ Passed"
		} else {
			msg = "✗ Failed"
		}
	} else if at.passing != result {
		if result {
			msg = "✓ Back to normal"
		} else {
			msg = "✗ Failed"
		}
	}
	return msg
}
