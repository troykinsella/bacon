package main

import (
	"fmt"
	"github.com/0xAX/notificator"
	"github.com/troykinsella/bacon/executor"
	"github.com/troykinsella/bacon/util"
	"github.com/troykinsella/bacon/watcher"
	"os"
	"text/template"
	"time"
	"strconv"
)

const (
	outputStatusFormat = "[{{ .timeStamp }}] {{ .colorStart }}{{ .statusSymbol }} {{ .status }}{{ .colorEnd }}"
	noOutputStatusFormat = "[{{ .timeSince }}] {{ .colorStart }}{{ .statusSymbol }} {{ .status }}{{ .colorEnd }}"

	statusRunning   = "Running"
	statusPassed    = "Passed"
	statusFailed    = "Failed"
	statusRecovered = "Back to normal"

	symbolRunning = "→"
	symbolPassed  = "✓"
	symbolFailed  = "✗"
)

type Bacon struct {
	w          *watcher.W
	e          *executor.E

	showOutput bool
	notify     bool
	statusChan chan *status
	n          *notificator.Notificator
}

type status struct {
	t time.Time
	running bool
	passing bool
}

func NewBacon(
	w *watcher.W,
	e *executor.E,
	showOutput bool,
	notify bool) *Bacon {

	statusChan := make(chan *status)

	b := &Bacon{
		w: w,
		e: e,

		showOutput: showOutput,
		notify:     notify,

		statusChan: statusChan,
		n:          newNotificator(),
	}

	go b.statusPrinter()

	return b
}

func (b *Bacon) statusPrinter() {

	var lastStatus *status

	for {
		select {
		case s := <- b.statusChan:
			lastStatus = s
			b.printStatus(s, false)

		case <- time.After(time.Second):
			if !b.showOutput {
				b.printStatus(lastStatus, true)
			}
		}
	}
}

func newNotificator() *notificator.Notificator {
	return notificator.New(notificator.Options{
		AppName: AppName,
	})
}

func (b *Bacon) Run() error {
	return b.w.Run(func(f string) {
		b.statusChan <- &status{
			t: time.Now(),
			running: true,
		}

		r := b.e.RunCommands(f, nil)

		b.statusChan <- &status{
			t: r.FinishedAt,
			passing: r.Passing,
		}

		b.pushNotification(r)
	})
}

func (b *Bacon) cls() {
	if !b.showOutput {
		util.Cls()
	}
}

func (b *Bacon) printStatus(s *status, repaint bool) {

	statusFmt := outputStatusFormat
	if !b.showOutput {
		statusFmt = noOutputStatusFormat
	}

	tpl, err := template.New("status").Parse(statusFmt + "\n")
	if err != nil {
		panic(err)
	}

	vars := b.statusVars(s)

	if s.running || s.passing {
		b.cls()
	}

	if !b.showOutput && repaint {
		fmt.Print("\033[1A                                \033[32D")
	}

	tpl.Execute(os.Stdout, vars)
}

func (b *Bacon) statusVars(s *status) map[string]string {

	now := time.Now()

	var status string
	var statusSymbol string
	var colorStart string
	var timeStamp string

	if s.running {
		status = statusRunning
		statusSymbol = symbolRunning
		colorStart = "\033[33m"
		timeStamp = now.Format("15:04:05")

	} else {
		if s.passing {
			status = statusPassed
			statusSymbol = symbolPassed
			colorStart = "\033[32m"
		} else {
			status = statusFailed
			statusSymbol = symbolFailed
			colorStart = "\033[31m"
		}

		timeStamp = s.t.Format("15:04:05")
	}

	return map[string]string{
		"showOutput":   strconv.FormatBool(b.showOutput),
		"status":       status,
		"statusSymbol": statusSymbol,
		"colorStart":   colorStart,
		"colorEnd":     "\033[0m",
		"timeStamp":    timeStamp,
		"timeSince":    round(now.Sub(s.t), time.Second).String(),
	}
}

func round(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
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
