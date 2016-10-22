package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

const (
	appName = "bacon"
	version = "0.0.2"

	command          = "c"
	commandLong      = command + ", cmd"
	passCommand      = "p"
	passCommandLong  = passCommand + ", pass"
	failCommand      = "f"
	failCommandLong  = failCommand + ", fail"
	debug            = "d"
	debugLong        = debug + ", debug"
	watch            = "w"
	watchLong        = watch + ", watch"
	watchExclude     = "e"
	watchExcludeLong = watchExclude + ", exclude"
	showOutput       = "o"
	showOutputLong   = showOutput + ", show-output"
	noNotify         = "no-notify"
)

func newOptions(c *cli.Context) (*Options, error) {
	cmds := c.StringSlice(command)
	if cmds == nil || len(cmds) == 0 {
		return nil, cli.NewExitError(command+" option required", 1)
	}

	passCmds := c.StringSlice(passCommand)
	failCmds := c.StringSlice(failCommand)
	includes := c.StringSlice(watch)
	excludes := c.StringSlice(watchExclude)
	showOut := c.Bool(showOutput)
	noNotify := c.Bool(noNotify)
	dbg := c.Bool(debug)

	return NewOptions(cmds,
		passCmds,
		failCmds,
		includes,
		excludes,
		showOut,
		!noNotify,
		dbg)
}

func newCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = appName
	app.Version = version
	app.Usage = "Watch files and run commands upon changes"
	app.Author = "Troy Kinsella"
	app.Action = func(c *cli.Context) error {
		opts, err := newOptions(c)
		if err != nil {
			return err
		}

		at := NewAutoTest(opts)
		err = at.Run()
		return err
	}
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  commandLong,
			Usage: "Shell `CMD` to execute. Required. Can be repeated.",
		},
		cli.StringSliceFlag{
			Name:  passCommandLong,
			Usage: "Run the `CMD` when tests pass. Can be repeated.",
		},
		cli.StringSliceFlag{
			Name:  failCommandLong,
			Usage: "Run the `CMD` when tests fail. Can be repeated.",
		},
		cli.BoolFlag{
			Name:  debugLong,
			Usage: "Enable " + app.Name + " debug output.",
		},
		cli.StringSliceFlag{
			Name:  watchLong,
			Usage: "Watch the path `GLOB`. Can be repeated. Defaults to './**/*'.",
		},
		cli.StringSliceFlag{
			Name:  watchExcludeLong,
			Usage: "Exclude path `GLOB` matches from being watched. Can be repeated. Defaults to '**/.git'.",
		},
		cli.BoolFlag{
			Name:  showOutputLong,
			Usage: "Enable command output.",
		},
		cli.BoolFlag{
			Name: noNotify,
			Usage: "Disable system notifications.",
		},
	}

	return app
}

func main() {
	cli.VersionFlag = cli.BoolFlag{
		Name:  "V, version",
		Usage: "print the version",
	}

	app := newCliApp()
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
