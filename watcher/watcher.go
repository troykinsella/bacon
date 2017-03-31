package watcher

import (
	"errors"
	"github.com/troykinsella/bacon/expander"
	"gopkg.in/fsnotify.v1"
	"strings"
	"time"
)

const eventThrottle = 100 * time.Millisecond

type W struct {
	e         *expander.E
	changed   ChangedFunc
	done      chan error
	fsWatcher *fsnotify.Watcher
	events    map[string]time.Time
}

type ChangedFunc func(f string)

func New(
	includes []string,
	excludes []string) (*W, error) {

	e := expander.New(includes, excludes)

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &W{
		e:         e,
		done:      make(chan error),
		fsWatcher: fsWatcher,
		events:    make(map[string]time.Time),
	}, nil
}

func (w *W) runDispatcher() {
	go func() {
		for {
			select {
			case event := <-w.fsWatcher.Events:
				n := extractName(event)
				s, err := w.e.Selected(n)
				if err != nil {
					w.done <- err
					break
				}
				if !s {
					continue
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					w.handleChange(n)
				}
			case err := <-w.fsWatcher.Errors:
				w.done <- err
				break
			}
		}
	}()
}

func (w *W) watchPaths(paths []string) error {
	for _, p := range paths {
		err := w.watchPath(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *W) watchPath(path string) error {
	err := w.fsWatcher.Add(path)
	return err
}

func (w *W) unwatchPath(path string) error {
	err := w.fsWatcher.Remove(path)
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

func (w *W) Run(changed ChangedFunc) error {
	w.changed = changed
	defer w.fsWatcher.Close()
	w.runDispatcher()

	dirs, err := w.e.BaseDirs()
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return errors.New("No paths to watch were matched")
	}

	w.watchPaths(dirs)
	w.changed("") // don't wait for a change

	return <-w.done
}

func (w *W) handleChange(f string) {
	last := w.events[f]
	if !last.IsZero() && time.Now().Sub(last) < eventThrottle {
		return
	}

	w.events[f] = time.Now()
	w.changed(f)
}
