package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

const (
	version = "0.0.1"

	arguments    = "a"
	debug        = "d"
	watch        = "w"
	watchExclude = "e"
	noTestOutput = "n"
	format       = "f"
)

func newOptions(c *cli.Context) (*Options, error) {
	args := c.String(arguments)
	if args == "" {
		return nil, cli.NewExitError("-"+arguments+" option required", 1)
	}

	includes := c.StringSlice(watch)
	excludes := c.StringSlice(watchExclude)
	noTO := c.Bool(noTestOutput)
	f := c.Bool(format)
	d := c.Bool(debug)

	return NewOptions(args, includes, excludes, !noTO, f, d)
}

func newCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = "autotest"
	app.Version = version
	app.Usage = "Watch go source files and run tests upon changes"
	app.Author = "Troy Kinsella"
	app.Action = func(c *cli.Context) error {
		opts, err := newOptions(c)
		if err != nil {
			return err
		}

		at := NewAutoTest(opts)
		err = at.Watch()
		return err
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  arguments,
			Usage: "go test `ARGS`. Required. Arguments are passed verbatim to 'go test' executions.",
		},
		cli.BoolFlag{
			Name:  debug,
			Usage: "Enable " + app.Name + " debug output.",
		},
		cli.StringSliceFlag{
			Name:  watch,
			Usage: "Watch the path `GLOB`. Can be repeated. Defaults to '$GOPATH/src/**/*.go'.",
		},
		cli.StringSliceFlag{
			Name:  watchExclude,
			Usage: "Exclude path `GLOB` matches from being watched. Can be repeated. Defaults to '**/.git/**'.",
		},
		cli.BoolFlag{
			Name:  noTestOutput,
			Usage: "No test output; only print summary messages.",
		},
		cli.BoolFlag{
			Name:  format,
			Usage: "Format a changed *.go file with 'go fmt' before executing tests.",
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
