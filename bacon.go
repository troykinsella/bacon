package main

import (
	"fmt"
	"github.com/0xAX/notificator"
	"github.com/troykinsella/bacon/executor"
	"github.com/troykinsella/bacon/util"
	"github.com/troykinsella/bacon/watcher"
	"time"
)

type Bacon struct {
	w *watcher.W
	e *executor.E

	clearScreen bool
	showOutput  bool
	notify      bool

	n *notificator.Notificator
}

func NewBacon(
	w *watcher.W,
	e *executor.E,
	clearScreen bool,
	showOutput bool,
	notify bool) *Bacon {

	return &Bacon{
		w: w,
		e: e,

		clearScreen: clearScreen,
		showOutput:  showOutput,
		notify:      notify,

		n: newNotificator(),
	}
}

func newNotificator() *notificator.Notificator {
	return notificator.New(notificator.Options{
		AppName: "bacon",
	})
}

func (b *Bacon) Run() error {
	return b.w.Run(func(f string) {
		b.cls()
		printRunning()

		r := b.e.RunCommands([]string{f})
		if r.Passing {
			b.cls()
		}

		printSummary(r.Passing)
		b.pushNotification(r)
	})
}

func (b *Bacon) cls() {
	if b.clearScreen && !b.showOutput {
		util.Cls()
	}
}

func printMessage(m string) {
	t := time.Now()
	fmt.Printf("[%s] %s\n", t.Format("15:04:05"), m)
}

func printRunning() {
	printMessage("\033[33m→ Running\033[0m")
}

func printSummary(pass bool) {
	var msg string
	if pass {
		msg = "\033[32m✓ Passed\033[0m"
	} else {
		msg = "\033[31m✗ Failed\033[0m"
	}
	printMessage(msg)
}

func (b *Bacon) pushNotification(r *executor.Result) {
	msg := b.notifyMessage(r)
	if msg != "" {
		b.n.Push("bacon", msg, "noop", notificator.UR_NORMAL)
	}
}

func (b *Bacon) notifyMessage(r *executor.Result) string {
	if !b.notify {
		return ""
	}

	var msg string
	if r.First {
		if r.Passing {
			msg = "✓ Passed"
		} else {
			msg = "✗ Failed"
		}
	} else if r.WasPassing != r.Passing {
		if r.Passing {
			msg = "✓ Back to normal"
		} else {
			msg = "✗ Failed"
		}
	}
	return msg
}
