package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/troykinsella/bacon/baconfile"
	"github.com/troykinsella/bacon/executor"
	"github.com/troykinsella/bacon/expander"
	"github.com/troykinsella/bacon/util"
	"github.com/troykinsella/bacon/watcher"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
	"strings"
)

const (
	AppName = "bacon"

	baconFile        = "b"
	baconFileLong    = baconFile + ", baconfile"
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
	shell            = "shell"

	defaultTarget = "default"
)

var (
	AppVersion = "0.0.0-dev.0"
)

func newExecutor(c *cli.Context) (*executor.E, error) {
	cmds := c.StringSlice(command)
	if cmds == nil || len(cmds) == 0 {
		return nil, cli.NewExitError(command+" option required", 1)
	}

	passCmds := c.StringSlice(passCommand)
	failCmds := c.StringSlice(failCommand)
	sh := c.String(shell)
	showOut := c.Bool(showOutput)

	e := executor.New(
		"",
		cmds,
		passCmds,
		failCmds,
		sh,
		"",
		showOut,
	)

	return e, nil
}

func newBacon(c *cli.Context) (*Bacon, error) {
	exp := expander.New(
		"",
		c.StringSlice(watch),
		c.StringSlice(watchExclude),
	)

	w, err := watcher.New(exp)
	if err != nil {
		return nil, err
	}

	exec, err := newExecutor(c)
	if err != nil {
		return nil, err
	}

	showOut := c.Bool(showOutput)
	noNotify := c.Bool(noNotify)

	b := NewBacon(
		w,
		exec,
		showOut,
		!noNotify,
	)
	return b, nil
}

func newListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "Print effective files to watch given inclusion and exclusion globs and exit.",
		Action: func(c *cli.Context) error {
			includes := c.StringSlice(watch)
			excludes := c.StringSlice(watchExclude)

			e := expander.New("", includes, excludes)
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

func readString(in *bufio.Scanner, msg string, def string) string {
	if def == "" {
		fmt.Printf("%s: ", msg)
	} else {
		fmt.Printf("%s [%s]: ", msg, def)
	}
	in.Scan()
	t := in.Text()
	if t == "" {
		t = def
	}
	return t
}

func readStringSlice(in *bufio.Scanner, msg string, def string, requireOne bool) []string {
	var result []string
	i := 0
	for {
		d := def
		if i > 0 {
			d = ""
		}

		s := readString(in, msg, d)
		if s == "" && requireOne {
			fmt.Println("At least one value is required")
			continue
		}

		if s == "" {
			break
		}

		result = append(result, s)

		if !readYesNo(in, "↪ Another list entry?", false) {
			break
		}

		i = i + 1
	}

	return result
}

func readYesNo(in *bufio.Scanner, msg string, def bool) bool {
	var defStr string
	if def {
		defStr = "Yn"
	} else {
		defStr = "yN"
	}
	fmt.Printf("%s [%s]: ", msg, defStr)
	in.Scan()
	t := in.Text()
	if t == "" {
		return def
	}
	return t == "y" || t == "Y"
}

func newInitCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Create a Baconfile by asking you questions.",
		Action: func(c *cli.Context) error {
			in := bufio.NewScanner(os.Stdin)

			targets := make(map[string]*baconfile.Target)

			for {
				var tName string
				for {
					tName = readString(in, "Target name", "default")
					if _, exists := targets[tName]; exists {
						fmt.Println("Already entered this target name. Enter a different name.")
						continue
					}
					break
				}

				dir := readString(in, "Working directory for patterns and commands", "")
				watch := readStringSlice(in, "Watch file pattern list", "**/*", true)
				cmd := readStringSlice(in, "Command list to run when files change", "", true)
				pass := readStringSlice(in, "Execute list when the commands pass", "", false)
				fail := readStringSlice(in, "Execute list when the commands fail", "", false)

				t := &baconfile.Target{
					Dir:     dir,
					Watch:   watch,
					Command: cmd,
					Pass:    pass,
					Fail:    fail,
				}

				targets[tName] = t

				if !readYesNo(in, "Create another target?", false) {
					break
				}
			}

			bf := baconfile.B{
				Targets: targets,
			}

			bytes, err := bf.Marshal()
			if err != nil {
				return err
			}

			if readYesNo(in, "View Baconfile preview?", false) {
				fmt.Print(string(bytes))
			}

			out := readString(in, "Baconfile path", "Baconfile")
			if !filepath.IsAbs(out) {
				cwd, err := os.Getwd()
				if err != nil {
					return err
				}
				out = filepath.Join(cwd, out)
			}

			if readYesNo(in, "Write Baconfile to "+out+"?", true) {
				err := os.WriteFile(out, bytes, 0644)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func newCommandCommand() *cli.Command {
	return &cli.Command{
		Name:  "command",
		Usage: "Execute commands once and exit. Useful for testing a command chain.",
		Action: func(c *cli.Context) error {
			e, err := newExecutor(c)
			if err != nil {
				return err
			}

			r := e.RunCommands("", nil)
			if !r.Passing {
				return cli.NewExitError("", 1)
			}

			return nil
		},
		Flags: newRunFlags(),
	}
}

func loadBaconfile(path string, required bool) (*baconfile.B, error) {
	exists, err := util.Exists(path)
	if err != nil {
		return nil, err
	}
	if !exists {
		if required {
			return nil, fmt.Errorf("baconfile not found: %s", path)
		}
		return nil, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	bf, err := baconfile.Unmarshal(bytes)
	if err != nil {
		return nil, err
	}

	return bf, nil
}

func findBaconfile(c *cli.Context) (*baconfile.B, error) {
	searchOrder := []string{
		c.String(baconFile),
		"Baconfile",
		"Baconfile.yml",
		"Baconfile.yaml",
		".Baconfile",
		".Baconfile.yml",
		".Baconfile.yaml",
	}

	for i, path := range searchOrder {
		if path == "" {
			continue
		}

		bf, err := loadBaconfile(path, i == 0)
		if err != nil {
			return nil, err
		}
		if bf != nil {
			return bf, nil
		}
	}

	return nil, errors.New("baconfile not found")
}

func newBaconForBaconfile(
	c *cli.Context,
	bc *baconfile.B,
	targetName string,
	args []string,
) (*Bacon, error) {
	if targetName == "" {
		targetName = defaultTarget
	}

	target := bc.Targets[targetName]
	if target == nil {
		// If the default target isn't found, just use the first one available
		if targetName == "default" {
			for tn, t := range bc.Targets {
				targetName = tn
				target = t
				break
			}
		}

		if target == nil {
			return nil, fmt.Errorf("baconfile target not found: %s", targetName)
		}
	}

	includes := injectArgs(target.Watch, args)
	excludes := injectArgs(target.Exclude, args)

	exp := expander.New(target.Dir, includes, excludes)
	w, err := watcher.New(exp)
	if err != nil {
		return nil, err
	}

	showOut := c.GlobalBool(showOutput)

	commands := injectArgs(target.Command, args)
	passCommands := injectArgs(target.Pass, args)
	failCommands := injectArgs(target.Fail, args)

	e := executor.New(
		targetName,
		commands,
		passCommands,
		failCommands,
		target.Shell,
		target.Dir,
		showOut,
	)

	noNotify := c.GlobalBool(noNotify)

	b := NewBacon(
		w,
		e,
		showOut,
		!noNotify,
	)
	return b, nil
}

func injectArgs(list []string, args []string) []string {
	result := list[:]
	for argIndex, arg := range args {
		for i, li := range result {
			result[i] = injectArg(li, argIndex+1, arg)
		}
	}
	return result
}

func injectArg(str string, index int, value string) string {
	varName := fmt.Sprintf("$%d", index)
	return strings.Replace(str, varName, value, -1)
}

func newRunCommand() *cli.Command {
	return &cli.Command{
		Name:      "run",
		Usage:     "Load configuration from a Baconfile target. The default target name is \"default\".",
		ArgsUsage: "[target] [target arguments]",
		Action: func(c *cli.Context) error {
			bf, err := findBaconfile(c)
			if err != nil {
				return err
			}

			var target string
			args := c.Args()
			if len(args) > 0 {
				target = args[0]
				args = args[1:]
			}

			b, err := newBaconForBaconfile(c, bf, target, args)
			if err != nil {
				return err
			}

			err = b.Run()
			if err != nil {
				return err
			}

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  baconFileLong,
				Usage: "The `PATH` to the Baconfile to load (default: Baconfile, Baconfile.yml, Baconfile.yaml)",
			},
		},
	}
}

func defCommands(app *cli.App) {
	app.Commands = []cli.Command{
		*newCommandCommand(),
		*newListCommand(),
		*newInitCommand(),
		*newRunCommand(),
	}
}

func newWatchFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringSliceFlag{
			Name:  watchLong,
			Usage: "Watch the path `GLOB`. Can be repeated. (default: \"**/*\")",
		},
		cli.StringSliceFlag{
			Name:  watchExcludeLong,
			Usage: "Exclude path `GLOB` matches from being watched. Can be repeated. (default: \"**/.*\")",
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
			Usage: "Run the `CMD` when commands pass. Can be repeated.",
		},
		cli.StringSliceFlag{
			Name:  failCommandLong,
			Usage: "Run the `CMD` when commands fail. Can be repeated.",
		},
		cli.StringFlag{
			Name:  shell,
			Usage: "The shell with which to interpret commands. (default: \"bash\")",
		},
	}
}

func newCliApp() *cli.App {
	app := cli.NewApp()
	app.Name = AppName
	app.Version = AppVersion
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
			Usage: "Enable command output",
		},
		cli.BoolFlag{
			Name:  noNotify,
			Usage: "Disable system notifications",
		},
	}

	app.Flags = append(app.Flags, newWatchFlags()...)
	app.Flags = append(app.Flags, newRunFlags()...)

	return app
}

func main() {
	app := newCliApp()
	err := app.Run(os.Args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
