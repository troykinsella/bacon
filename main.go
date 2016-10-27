package main

import (
	"fmt"
	"github.com/troykinsella/bacon/executor"
	"github.com/troykinsella/bacon/watcher"
	"github.com/urfave/cli"
	"os"
)

const (
	appName = "bacon"
	version = "0.0.3"

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

func newExecutor(c *cli.Context, cls bool) (*executor.E, error) {
	cmds := c.StringSlice(command)
	if cmds == nil || len(cmds) == 0 {
		return nil, cli.NewExitError(command+" option required", 1)
	}

	passCmds := c.StringSlice(passCommand)
	failCmds := c.StringSlice(failCommand)

	showOut := c.Bool(showOutput)

	e := executor.New(cmds,
		passCmds,
		failCmds,
		cls,
		showOut)

	return e, nil
}

func newWatcher(c *cli.Context) (*watcher.W, error) {
	includes := c.StringSlice(watch)
	excludes := c.StringSlice(watchExclude)
	dbg := c.Bool(debug)

	w, err := watcher.New(includes, excludes, dbg)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func newBacon(c *cli.Context) (*Bacon, error) {
	w, err := newWatcher(c)
	if err != nil {
		return nil, err
	}

	e, err := newExecutor(c, true)
	if err != nil {
		return nil, err
	}

	showOut := c.Bool(showOutput)
	noNotify := c.Bool(noNotify)

	b := NewBacon(w, e, true, showOut, !noNotify)
	return b, nil
}

func newRunCommand() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Execute commands once. Useful for testing a command chain.",
		Action: func(c *cli.Context) error {
			e, err := newExecutor(c, false)
			if err != nil {
				return err
			}

			r := e.RunCommands(nil)
			if !r.Passing {
				return cli.NewExitError("", 1)
			}

			return nil
		},
		Flags: newRunFlags(),
	}
}

func defCommands(app *cli.App) {
	app.Commands = []cli.Command{
		*newRunCommand(),
	}
}

func newWatchFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringSliceFlag{
			Name:  watchLong,
			Usage: "Watch the path `GLOB`. Can be repeated. Defaults to './**/*'.",
		},
		cli.StringSliceFlag{
			Name:  watchExcludeLong,
			Usage: "Exclude path `GLOB` matches from being watched. Can be repeated. Defaults to '**/.git'.",
		},
	}
}

func newRunFlags() []cli.Flag {
	return []cli.Flag{
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
	}
}

func newCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = appName
	app.Version = version
	app.Usage = "Watch files and run commands upon changes"
	app.Author = "Troy Kinsella"
	app.Action = func(c *cli.Context) error {
		b, err := newBacon(c)
		if err != nil {
			return err
		}
		err = b.Run()
		return err
	}

	defCommands(app)

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  debugLong,
			Usage: "Enable " + app.Name + " debug output.",
		},
		cli.BoolFlag{
			Name:  showOutputLong,
			Usage: "Enable command output.",
		},
		cli.BoolFlag{
			Name:  noNotify,
			Usage: "Disable system notifications.",
		},
	}

	app.Flags = append(app.Flags, newWatchFlags()...)
	app.Flags = append(app.Flags, newRunFlags()...)

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
