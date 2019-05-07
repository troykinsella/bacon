package watcher

import (
	"errors"
	"github.com/troykinsella/bacon/expander"
	"gopkg.in/fsnotify.v1"
	"os"
	"time"
)

type W struct {
	exp       *expander.E
	changed   ChangedFunc
	done      chan error
	fsWatcher *fsnotify.Watcher
	lastMods  map[string]time.Time
}

type ChangedFunc func(f string)

func New(exp *expander.E) (*W, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &W{
		exp:       exp,
		done:      make(chan error),
		fsWatcher: fsWatcher,
		lastMods:  make(map[string]time.Time),
	}, nil
}

func (w *W) acceptEvent(path string) (bool, error) {
	s, err := w.exp.Selected(path)
	if err != nil {
		return false, err
	}
	if !s {
		return false, nil
	}

	stat, err := os.Stat(path)
	if err != nil {
		delete(w.lastMods, path)
		return false, nil // ignore
	}

	lastMod, ok := w.lastMods[path]
	curMod := stat.ModTime()
	w.lastMods[path] = curMod
	if ok && lastMod == curMod {
		return false, nil
	}

	return true, nil
}

func (w *W) changeWatcher() {
	for {
		select {
		case event := <-w.fsWatcher.Events:
			ok, err := w.acceptEvent(event.Name)
			if err != nil {
				w.done <- err
				break
			}
			if !ok {
				continue
			}

			go w.changed(event.Name)

		case err := <-w.fsWatcher.Errors:
			w.done <- err
			break
		}
	}
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

func (w *W) Run(changed ChangedFunc) error {
	w.changed = changed
	defer w.fsWatcher.Close()
	go w.changeWatcher()

	dirs, err := w.exp.BaseDirs()
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
