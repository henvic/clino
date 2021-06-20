// Package clino provides a simple way to create CLI (command-line interface) tools.
//
// You can create commands to use with this package by implementing its interfaces.
// It supports the Unix -flag style.
//
// The Command interface contains only a name.
// However, if you try to run a command that doesn't implement any of the
// Runnable, Longer, Parent, or Footer interfaces,
// you are going to get a "missing implementation" error message.
//
// For working with flags, you need to implement the FlagSet interface to a given command.
// If you need global flags, you can do so by defining Program.GlobalFlags.
// You can use it for a -verbose, -config, or other application-wide state flags.
// In example/complex you can see how to use global flags easily.
package clino

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Command contains the minimal interface for a command: its name (usage).
//
// You usually want to implement the Runnable interface, except for
// help-only commands, or when your command has subcommands (implements Parent).
type Command interface {
	Name() string
}

// Shorter description of a command to show in the "help" output on a list of commands.
type Shorter interface {
	Short() string
}

// Runnable commands are commands that implement the Run function, and you can run it from the command-line.
// It should receive a context and the command arguments, after parsing any flags.
// A context is required as we want cancelation to be a first-class citizen.
// You can rely on the context for canceling long tasks during tests.
type Runnable interface {
	Run(ctx context.Context, args ...string) error
}

// FlagSet you want to use on your command.
// 	// Flags of the "hello" command.
// 	func (hc *HelloCommand) Flags(flags *flag.FlagSet) {
//		flags.StringVar(&hc.name, "name", "World", "your name")
// 	}
// You need to implement a Flags function like shown and set any flags you want your commands to parse.
type FlagSet interface {
	Flags(flags *flag.FlagSet)
}

// PersistentFlagSet is similar to FlagSet, but flags are inherited by the next commands.
// 	// PersistentFlags of the "main" command.
// 	func (mc *MainCommand) PersistentFlags(flags *flag.FlagSet) {
//		flags.BoolVar(&hc.verbose, "verbose", false, "verbose mode")
// 	}
// You need to implement a Flags function like shown and set any flags you want your commands to parse.
type PersistentFlagSet interface {
	PersistentFlags(flags *flag.FlagSet)
}

// Longer description or help message for your command.
// The help command prints the returned value of the Long function as the "help" output of a command.
type Longer interface {
	Long() string
}

// Footer of a command shown in the "help <command>" output.
// It is useful for things like printing examples.
type Footer interface {
	Foot() string
}

// Parent contains all subcommands of a given command.
type Parent interface {
	Commands() []Command
}

// Program you want to run.
//
// You should call the Run function, passing the context, root command, and process arguments.
type Program struct {
	// Root command is the entrypoint of the program.
	Root Command

	// GlobalFlags are flags available to all commands.
	//
	// Deprecated: Use PersistentFlags instead.
	GlobalFlags func(flags *flag.FlagSet)

	// Output is the default output function to the application.
	//
	// If not set when calling Run, os.Stdout is set.
	// You probably only want to set this for testing.
	Output io.Writer

	fs *flag.FlagSet
}

// Run program by processing arguments and executing the invoked command.
//
// Context is passed down to the command to simplify testing and cancelation.
// Arguments should be the process arguments (os.Args[1:]...) when you call it from main().
//
// Example:
// p := clino.Program{
// 	Root: &RootCommand{},
// }
// if err := p.Run(context.Background(), os.Args[1:]...); err != nil {
// 	fmt.Fprintf(os.Stderr, "%+v\n", err)
// 	os.Exit(clino.ExitCode(err))
// }
func (p *Program) Run(ctx context.Context, args ...string) error {
	if p.Output == nil {
		p.Output = os.Stdout
	}
	if p.Root == nil {
		panic("root command not implemented")
	}
	checkDuplicated(p.Root, []string{p.Root.Name()})
	p.fs = flag.NewFlagSet("", flag.ContinueOnError)
	p.fs.SetOutput(ioutil.Discard) // skip printing flags -help when parsing flags fail.
	if p.GlobalFlags != nil {
		p.GlobalFlags(p.fs)
	}
	return p.runCommand(ctx, args)
}

// checkDuplicated is supposed to be called initially with the root command and check the children implementations, recursively.
func checkDuplicated(cmd Command, trail []string) {
	p, ok := cmd.(Parent)
	if !ok {
		return
	}
	var m = map[string]struct{}{}
	for _, c := range p.Commands() {
		name, cmdtrail := c.Name(), append(trail, c.Name())
		if _, ok := m[name]; ok {
			panic("command implemented multiple times: '" + strings.Join(cmdtrail, " ") + "'")
		}
		m[name] = struct{}{}
		checkDuplicated(c, cmdtrail)
	}
}

func isRunnable(cmd Command) bool {
	_, ok := cmd.(Runnable)
	return ok
}

func commandNotFound(binary string, trail []string) error {
	trail = append([]string{binary}, trail...)
	return fmt.Errorf("unknown command: '%v'", strings.Join(trail, " "))
}

func (p *Program) loadCommand(ctx context.Context, args []string) []Command {
	commands := getSubcommands(p.Root)
	cmdArgs := getCommandArgs(args)
	return p.walkCommand(commands, cmdArgs)
}

func skipHelpCommand(args []string) []string {
	if len(args) != 0 && args[0] == "help" {
		return args[1:]
	}
	return args
}

func (p *Program) runCommand(ctx context.Context, args []string) error {
	trail := p.loadCommand(ctx, skipHelpCommand(args))
	cmd := trail[len(trail)-1]

	for _, c := range trail {
		if f, ok := c.(PersistentFlagSet); ok && f != nil {
			f.PersistentFlags(p.fs)
		}
	}
	if f, ok := cmd.(FlagSet); ok && f != nil {
		f.Flags(p.fs)
	}
	if (len(args) == 0 && !isRunnable(p.Root)) || (len(args) != 0 && args[0] == "help") {
		return p.runHelp(ctx, args)
	}
	if r, ok := cmd.(Runnable); ok && r != nil {
		err := p.fs.Parse(args[len(trail)-1:])
		if err == flag.ErrHelp {
			return p.runHelp(ctx, args)
		}
		if err != nil {
			return err
		}
		return r.Run(ctx, p.fs.Args()...)
	}
	return p.runHelp(ctx, args)
}

func (p *Program) runHelp(ctx context.Context, args []string) error {
	if len(args) >= 1 && args[0] == "help" {
		args = args[1:]
	}
	trail := p.walkCommand(getSubcommands(p.Root), getCommandArgs(args))
	cmd := trail[len(trail)-1]

	var breadcrumb []string
	for _, c := range trail {
		breadcrumb = append(breadcrumb, c.Name())
	}
	breadcrumb = breadcrumb[1:]

	h := &helper{
		Output:   p.Output,
		Commands: getSubcommands(cmd),
		binary:   p.Root.Name(),
		trail:    breadcrumb,
		args:     args,
		fs:       p.fs,
	}
	if l, ok := cmd.(Longer); ok && l != nil {
		h.Long = l.Long
	}
	if f, ok := cmd.(Footer); ok && f != nil {
		h.Foot = f.Foot
	}

	p.setUsableHelp(cmd, h)

	return h.Run(ctx)
}

// setUsableHelp is used to only print help for flags and 'usage' message
// if command has subcommands or is runnable.
func (p *Program) setUsableHelp(cmd Command, h *helper) {
	_, h.runnable = cmd.(Runnable)
	_, parent := cmd.(Parent)
	h.usable = h.runnable || parent
}

func getCommand(commands []Command, name string) (cmd Command, ok bool) {
	for _, c := range commands {
		if name == c.Name() {
			return c, true
		}
	}
	return
}

// walkCommand is similar to getCommand, but recursive and it stops
// when it can't find any further command following the path.
// The returned trail value is the "breadcrumb" for the command.
func (p *Program) walkCommand(commands []Command, names []string) (trail []Command) {
	trail = append(trail, p.Root)
	current := commands
	for _, name := range names {
		c, next := getCommand(current, name)
		if !next {
			return
		}
		trail = append(trail, c)
		current = getSubcommands(c)
	}
	return
}

func getCommandArgs(args []string) (out []string) {
	if len(args) == 0 {
		return
	}
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") { // stop on first flag
			return
		}
		out = append(out, arg)
	}
	return
}

func getSubcommands(cmd Command) []Command {
	if p, ok := cmd.(Parent); ok && p != nil {
		return p.Commands()
	}
	return []Command{}
}
