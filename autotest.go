package main

import (
	"errors"
	"fmt"
	"github.com/0xAX/notificator"
	"github.com/bmatcuk/doublestar"
	"gopkg.in/fsnotify.v1"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Options struct {
	targets    []string
	cmd        []string
	testOutput bool
	goFmt      bool
	debug      bool
}

func NewOptions(args string,
	includes []string,
	excludes []string,
	testOutput bool,
	goFmt bool,
	debug bool) (*Options, error) {

	expTargets, err := expandTargets(includes, excludes)
	if err != nil {
		return nil, err
	}

	if len(expTargets) == 0 {
		return nil, errors.New("No paths to watch were matched")
	}

	if debug {
		fmt.Println("Watching targets:")
		for _, t := range expTargets {
			fmt.Println(t)
		}
	}

	return &Options{
		targets:    expTargets,
		cmd:        createCommand(args),
		testOutput: testOutput,
		goFmt:      goFmt,
		debug:      debug,
	}, nil
}

type AutoTest struct {
	opts    *Options
	n       *notificator.Notificator
	first   bool
	passing bool
}

func NewAutoTest(o *Options) *AutoTest {
	return &AutoTest{
		opts:    o,
		n:       newNotificator(),
		first:   true,
		passing: true,
	}
}

func newNotificator() *notificator.Notificator {
	return notificator.New(notificator.Options{
		AppName: "autotest",
	})
}

func goSourceDir() string {
	gp, _ := os.LookupEnv("GOPATH")
	if gp == "" {
		panic(errors.New("GOPATH must be set"))
	}
	return path.Join(gp, "src")
}

func isDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	s, err := f.Stat()
	if err != nil {
		return false, err
	}
	return s.IsDir(), nil
}

func excluded(path string, excludes []string) (bool, error) {
	for _, e := range excludes {
		m, err := doublestar.PathMatch(e, path)
		if err != nil {
			return false, err
		}
		if m {
			dir, err := isDir(path)
			if err != nil {
				return false, err
			}
			if dir {
				return false, nil
			}

			return true, nil
		}
	}
	return false, nil
}

func normalizeGlobs(globs []string, defalt string) []string {
	if globs == nil || len(globs) == 0 {
		globs = []string{defalt}
	}

	goSrc := goSourceDir()
	for i, g := range globs {
		if strings.HasPrefix(g, "/") {
			globs[i] = g
		} else {
			globs[i] = filepath.Join(goSrc, g)
		}
	}

	return globs
}

func expandTargets(includes []string, excludes []string) ([]string, error) {
	includes = normalizeGlobs(includes, "**/*.go")
	excludes = normalizeGlobs(excludes, "**/.git/**")

	result := []string{}

	for _, inc := range includes {
		matches, err := doublestar.Glob(inc)
		if err != nil {
			return nil, err
		}

		for _, m := range matches {
			e, err := excluded(m, excludes)
			if err != nil {
				return nil, err
			}

			if !e {
				result = append(result, m)
			}
		}
	}

	return result, nil
}

func createCommand(args string) []string {
	c := []string{
		"test",
	}

	parts := strings.Split(args, " ")
	for _, arg := range parts {
		c = append(c, arg)
	}

	return c
}

func (at *AutoTest) addTargets(w *fsnotify.Watcher) error {
	for _, t := range at.opts.targets {
		err := w.Add(t)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractName(e fsnotify.Event) string {
	n := e.Name

	// Lame-ass JetBrains hack: remove a "___jb_tmp___" suffix
	const sfx = "___jb_tmp___"
	if strings.HasSuffix(n, sfx) {
		n = n[0 : len(n)-len(sfx)]
	}

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
				if event.Op&fsnotify.Write == fsnotify.Write {
					n := extractName(event)

					if at.opts.debug {
						fmt.Printf("Changed: %s\n", n)
					}

					at.handleChange(n)
				}
			case err := <-w.Errors:
				done <- err
			}
		}
	}()

	return w, nil
}

func (at *AutoTest) Watch() error {
	done := make(chan error)
	watcher, err := at.newWatcher(done)
	if err != nil {
		return err
	}
	defer watcher.Close()

	at.addTargets(watcher)
	at.runTests() // don't wait for a change

	return <-done
}

func (at *AutoTest) goFmtFile(f string) bool {
	if !at.opts.goFmt {
		return false
	}

	if !strings.HasSuffix(f, ".go") {
		return false
	}

	fmt.Printf("Formatting: %s\n", f)

	cmd := exec.Command("go", "fmt", f)
	cmd.Run()
	return true
}

func (at *AutoTest) handleChange(f string) {
	if at.goFmtFile(f) {
		// If we formatted the file, it will trigger another write event,
		// so we can just drop this one.
		//return Hmm I'm not getting subsequent change events
	}

	// TODO: Throttle test runs here

	at.runTests()
}

func (at *AutoTest) runTests() {
	cmd := exec.Command("go", at.opts.cmd...)
	if at.opts.testOutput {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	result := err == nil

	at.printSummary(result)
	at.notify(result)

	at.first = false
	at.passing = result
}

func (at *AutoTest) printSummary(pass bool) {
	var msg string
	if pass {
		msg = "\033[32m✓\033[0m Tests passed"
	} else {
		msg = "\033[31m✗\033[0m Tests failed"
	}

	t := time.Now()
	fmt.Printf("[%s] %s\n", t.Format("15:04:05"), msg)
}

func (at *AutoTest) notify(result bool) {
	msg := at.notifyMessage(result)
	if msg != "" {
		at.n.Push("go test", msg, "noop", notificator.UR_NORMAL)
	}
}

func (at *AutoTest) notifyMessage(result bool) string {
	var msg string
	if at.first {
		if result {
			msg = "✓ Tests passing"
		} else {
			msg = "✗ Tests failing"
		}
	} else if at.passing != result {
		if result {
			msg = "✓ Tests back to normal"
		} else {
			msg = "✗ Tests failing"
		}
	}
	return msg
}
