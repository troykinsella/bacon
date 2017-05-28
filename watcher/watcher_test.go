package watcher

import (
	"os/exec"
	"testing"
	"time"
	"github.com/troykinsella/bacon/expander"
)

func TestW_Run(t *testing.T) {
	exp := expander.New("", []string{"testdata/foo"}, []string{})
	w, err := New(exp)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}

	done := make(chan bool)

	go w.Run(func(f string) {
		done <- true
	})

	// Ensure called right away
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Error("Initial callback timed out")
		return
	}

	// Touch watched file
	err = exec.Command("sh", "-c", "echo 1 > testdata/foo").Run()
	if err != nil {
		t.Errorf("File change error: %s", err.Error())
		return
	}

	// Ensure called in reaction to touched file
	select {
	case <-done:
		// Ensure not called again
		select {
		case <-done:
			t.Error("Called back twice")
		case <-time.After(eventThrottle):
		}
	case <-time.After(1 * time.Second):
		t.Error("Watch callback timed out")
	}
}
