package clino

import (
	"context"
	"flag"
	"fmt"
	"time"
)

type simpleCommand struct {
	ran  bool
	args []string
	name string
}

// Name of the application.
func (sc *simpleCommand) Name() string {
	return "simple"
}

// Long description of the application.
func (sc *simpleCommand) Long() string {
	return "Example application."
}

// Flags of the command.
func (sc *simpleCommand) Flags(flags *flag.FlagSet) {
	flags.StringVar(&sc.name, "name", "World", "your name")
}

// Run command.
func (sc *simpleCommand) Run(ctx context.Context, args ...string) error {
	sc.ran = true
	sc.args = args
	fmt.Printf("Hello, %s!\n", sc.name)
	return nil
}

type persistentFlagsCommand struct {
	simpleCommand
	verbose bool
}

// PersistentFlags of the command.
func (pc *persistentFlagsCommand) PersistentFlags(flags *flag.FlagSet) {
	flags.BoolVar(&pc.verbose, "verbose", false, "verbose mode")
	var persistentflag string
	flags.StringVar(&persistentflag, "persistentflag", "none", "persistent flag")
}

// rootCommand is an application with multiple commands
type rootCommand struct{}

// Commands for the application.
func (rc *rootCommand) Commands() []Command {
	return []Command{
		&notRunnableCommand{},
		&unimplementedCommand{},
	}
}

// Name of the application.
func (rc *rootCommand) Name() string {
	return "app"
}

// Long description of the application.
func (rc *rootCommand) Long() string {
	return "Example application."
}

// Foot containing examples.
func (rc *rootCommand) Foot() string {
	return `Example: add anything here.

If you like this library, let me know!`
}

// anotherCommand is an application with multiple commands
type anotherCommand struct{}

// Commands for the application.
func (rc *anotherCommand) Commands() []Command {
	return []Command{
		&notRunnableCommand{},
		&unimplementedCommand{},
		&simpleCommand{},
	}
}

// Name of the application.
func (rc *anotherCommand) Name() string {
	return "app"
}

// notRunnableCommand is a command compromised only of help text.
type notRunnableCommand struct{}

// Name of the command.
func (nrc *notRunnableCommand) Name() string {
	return "not-runnable"
}

// Short description of the command.
func (nrc *notRunnableCommand) Short() string {
	return `command containing a help topic`
}

func (nrc *notRunnableCommand) Long() string {
	return `This is a not so long,
multiline help topic.`
}

// unimplementedCommand is a command that doesn't implement any useful interfaces
// like Runnable, Longer, or Parent interfaces.
type unimplementedCommand struct{}

func (uic *unimplementedCommand) Name() string {
	return "unimplemented"
}

// badRootCommand contains two commands with the same name.
// It will panic if initialized.
type badRootCommand struct{}

// Name for the bad root command.
func (brc *badRootCommand) Name() string {
	return "bad"
}

// Commands of the bad root command.
// We are simply initializing the same command twice.
func (brc *badRootCommand) Commands() []Command {
	return []Command{
		&simpleCommand{},
		&simpleCommand{},
	}
}

// rootCommandWithFlags is an application with multiple commands
type rootCommandWithFlags struct {
	planet   string
	deadline time.Duration
	verbose  bool
	open     bool
	nodes    *int
	custom   customFlag
}

// Name of the application.
func (rcf *rootCommandWithFlags) Name() string {
	return "cmd"
}

// Commands for the application.
func (rcf *rootCommandWithFlags) Commands() []Command {
	return []Command{
		&notRunnableCommand{},
		&innerCommand{},
	}
}

func (rcf *rootCommandWithFlags) Run(ctx context.Context, args ...string) error {
	fmt.Println("custom string: " + rcf.custom.String())
	return nil
}

type rootCommandWithFlagsAndPersistentFlags struct {
	rootCommandWithFlags
}

func (rfp *rootCommandWithFlagsAndPersistentFlags) PersistentFlags(flags *flag.FlagSet) {
	var persistentflag string
	flags.StringVar(&persistentflag, "persistentflag", "none", "persistent flag")
}

type customFlag struct {
	value string
}

func (c *customFlag) String() string { return c.value }

func (c *customFlag) Set(v string) error {
	c.value = v
	return nil
}

// unusedFlag is defined to cover the isZeroValue function when not dealing with a pointer.
// It could theoretically be useful for discontinued or obsolete flags after deprecation, for example.
type unusedBoolFlag struct{}

func (f unusedBoolFlag) String() string     { return "" }
func (f unusedBoolFlag) Set(v string) error { return nil }
func (f unusedBoolFlag) IsBoolFlag() bool   { return true }

// Long description of the application.
func (rcf *rootCommandWithFlags) Flags(flags *flag.FlagSet) {
	flags.StringVar(&rcf.planet, "planet", "Earth", "name of the planet")
	flags.DurationVar(&rcf.deadline, "deadline", 5*time.Second, "deadline for the operation")
	flags.BoolVar(&rcf.verbose, "verbose", false, "show more information")
	flags.BoolVar(&rcf.open, "open", true, "open link in the browser")
	rcf.nodes = flags.Int("nodes", 1, "number of nodes")
	flags.Var(&rcf.custom, "custom", "custom flag test")
	flags.Var(unusedBoolFlag{}, "unused", "unused bool flag")
}

type innerCommand struct{}

func (s *innerCommand) Name() string {
	return "inner"
}

func (s *innerCommand) Commands() []Command {
	return []Command{
		&notRunnableCommand{},
		&simpleCommand{},
	}
}
