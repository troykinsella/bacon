package main

import (
	"fmt"
	"github.com/0xAX/notificator"
	"github.com/troykinsella/bacon/executor"
	"github.com/troykinsella/bacon/util"
	"github.com/troykinsella/bacon/watcher"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

const (
	defaultSummaryFormat = "[{{ .timeStamp }}] {{ .colorStart }}{{ .statusSymbol }} {{ .status }}{{ .colorEnd }}"

	statusRunning   = "Running"
	statusPassed    = "Passed"
	statusFailed    = "Failed"
	statusRecovered = "Back to normal"

	symbolRunning = "→"
	symbolPassed  = "✓"
	symbolFailed  = "✗"
)

type Bacon struct {
	w *watcher.W
	e *executor.E

	showOutput bool
	notify     bool
	summaryFmt string

	n *notificator.Notificator
}

func NewBacon(
	w *watcher.W,
	e *executor.E,
	showOutput bool,
	notify bool,
	summaryFmt string) *Bacon {

	if summaryFmt == "" {
		summaryFmt = defaultSummaryFormat
	}

	return &Bacon{
		w: w,
		e: e,

		showOutput: showOutput,
		notify:     notify,
		summaryFmt: summaryFmt,

		n: newNotificator(),
	}
}

func newNotificator() *notificator.Notificator {
	return notificator.New(notificator.Options{
		AppName: AppName,
	})
}

func (b *Bacon) Run() error {
	return b.w.Run(func(f string) {
		b.cls()
		b.printSummary(nil, f, nil)

		start := time.Now()
		r := b.e.RunCommands([]string{f})
		if r.Passing {
			b.cls()
		}

		b.printSummary(&start, f, r)
		b.pushNotification(r)
	})
}

func (b *Bacon) cls() {
	if !b.showOutput {
		util.Cls()
	}
}

func (b *Bacon) printSummary(start *time.Time, changed string, r *executor.Result) {
	tpl, err := template.New("summary").Parse(b.summaryFmt + "\n")
	if err != nil {
		panic(err)
	}

	vars := summaryVars(start, changed, r)
	tpl.Execute(os.Stdout, vars)
}

func summaryVars(start *time.Time, changedFile string, r *executor.Result) map[string]string {

	end := time.Now()

	var changedDir string
	var status string
	var statusSymbol string
	var colorStart string
	var duration string

	if r == nil {
		status = statusRunning
		statusSymbol = symbolRunning
		colorStart = "\033[33m"
	} else {
		if r.Passing {
			status = statusPassed
			statusSymbol = symbolPassed
			colorStart = "\033[32m"
		} else {
			status = statusFailed
			statusSymbol = symbolFailed
			colorStart = "\033[31m"
		}

		duration = end.Sub(*start).String()
	}

	if changedFile != "" {
		changedDir = filepath.Dir(changedFile)
	}

	return map[string]string{
		"changedDir":   changedDir,
		"changedFile":  changedFile,
		"status":       status,
		"statusSymbol": statusSymbol,
		"colorStart":   colorStart,
		"colorEnd":     "\033[0m",
		"timeStamp":    end.Format("15:04:05"),
		"duration":     duration,
	}
}

func (b *Bacon) pushNotification(r *executor.Result) {
	if !b.notify {
		return
	}

	msg := b.notifyMessage(r)
	if msg != "" {
		b.n.Push(AppName, msg, "noop", notificator.UR_NORMAL)
	}
}

func (b *Bacon) notifyMessage(r *executor.Result) string {
	var msg string
	if r.First {
		if r.Passing {
			msg = fmt.Sprintf("%s %s", symbolPassed, statusPassed)
		} else {
			msg = fmt.Sprintf("%s %s", symbolFailed, statusFailed)
		}
	} else if r.WasPassing != r.Passing {
		if r.Passing {
			msg = fmt.Sprintf("%s %s", symbolPassed, statusRecovered)
		} else {
			msg = fmt.Sprintf("%s %s", symbolFailed, statusFailed)
		}
	}
	return msg
}
