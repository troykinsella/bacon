package main

import (
	"fmt"
	"github.com/troykinsella/bacon/executor"
	"github.com/troykinsella/bacon/expander"
	"github.com/troykinsella/bacon/watcher"
	"github.com/urfave/cli"
	"os"
)

const (
	appName = "bacon"
	version = "0.0.5"

	command          = "c"
	commandLong      = command + ", cmd"
	passCommand      = "p"
	passCommandLong  = passCommand + ", pass"
	failCommand      = "f"
	failCommandLong  = failCommand + ", fail"
	watch            = "w"
	watchLong        = watch + ", watch"
	watchExclude     = "e"
	watchExcludeLong = watchExclude + ", exclude"
	showOutput       = "o"
	showOutputLong   = showOutput + ", show-output"
	noNotify         = "no-notify"
	summaryFormat    = "summary-format"
	shell            = "shell"
)

func newExecutor(c *cli.Context, cls bool) (*executor.E, error) {
	cmds := c.StringSlice(command)
	if cmds == nil || len(cmds) == 0 {
		return nil, cli.NewExitError(command+" option required", 1)
	}

	passCmds := c.StringSlice(passCommand)
	failCmds := c.StringSlice(failCommand)
	sh := c.String(shell)
	showOut := c.Bool(showOutput)

	e := executor.New(cmds,
		passCmds,
		failCmds,
		sh,
		cls,
		showOut)

	return e, nil
}

func newExpander(c *cli.Context) (*expander.E, error) {
	includes := c.StringSlice(watch)
	excludes := c.StringSlice(watchExclude)
	e := expander.New(includes, excludes)
	return e, nil
}

func newWatcher(c *cli.Context) (*watcher.W, error) {
	includes := c.StringSlice(watch)
	excludes := c.StringSlice(watchExclude)

	w, err := watcher.New(includes, excludes)
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
	sumFmt := c.String(summaryFormat)

	b := NewBacon(w,
		e,
		showOut,
		!noNotify,
		sumFmt)
	return b, nil
}

func newListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Print effective files to watch given inclusion and exclusion globs and exit.",
		Action: func(c *cli.Context) error {
			e, err := newExpander(c)
			if err != nil {
				return err
			}

			list, err := e.List()
			if err != nil {
				return err
			}

			for _, f := range list {
				fmt.Println(f)
			}
			return nil
		},
		Flags: newWatchFlags(),
	}
}

func newRunCommand() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Execute commands once and exit. Useful for testing a command chain.",
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
		*newListCommand(),
		*newRunCommand(),
	}
}

func newWatchFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringSliceFlag{
			Name:  watchLong,
			Usage: "Watch the path `GLOB`. Can be repeated. Defaults to '**/*'.",
		},
		cli.StringSliceFlag{
			Name:  watchExcludeLong,
			Usage: "Exclude path `GLOB` matches from being watched. Can be repeated. Defaults to '**/.*'.",
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
		cli.StringFlag{
			Name: shell,
			Usage: "The shell with which to interpret commands. Default 'bash'.",
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
			Name:  showOutputLong,
			Usage: "Enable command output.",
		},
		cli.BoolFlag{
			Name:  noNotify,
			Usage: "Disable system notifications.",
		},
		cli.StringFlag{
			Name:  summaryFormat,
			Usage: "Go template string for custom summary lines.",
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
