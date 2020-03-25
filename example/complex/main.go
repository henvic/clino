package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/henvic/clino"
)

// State of the application holds flags that persists between the
// root command and its offspring, plus any other value you would like it to.
type State struct {
	Verbose bool
}

// Flags implements the clino.FlagSet.
func (s *State) Flags(flags *flag.FlagSet) {
	flags.BoolVar(&s.Verbose, "verbose", false, "show more information")
}

// RootCommand is the main command for the application.
type RootCommand struct {
	State *State
}

// Commands for the application.
func (rc *RootCommand) Commands() []clino.Command {
	return []clino.Command{
		&AboutCommand{},
		&HelloCommand{
			State: rc.State,
		},
	}
}

// Name of the application.
func (rc *RootCommand) Name() string {
	return "app"
}

// Long description of the application.
func (rc *RootCommand) Long() string {
	return "Example application."
}

// Foot containing examples.
func (rc *RootCommand) Foot() string {
	return `Example: add anything here.

If you like this library, let me know!`
}

// AboutCommand explains what this library is about.
type AboutCommand struct{}

// Name of the "about" command.
func (ac *AboutCommand) Name() string {
	return "about"
}

// Long topic about this command.
func (ac *AboutCommand) Long() string {
	return `This library offers a different approach for creating CLI tools.

Instead of instantiating structs to create multiple commands with Go,
you should implement code that implements a few interfaces.
This library was created as an experiment to validate this idea.

One advantage is that it is easier to mock and test the essential parts.
One drawback is that you have to implement the exact interface for it to match and work.

Starting with structs is easy because your code editor auto-completion should kick.
However, you usually end up creating globals for things like flags, making testing a little harder.

With this interface approach, you have to check the code and documentation to understand what you need.
`
}

// HelloCommand is used to print a "Hello, World!" message.
type HelloCommand struct {
	State *State
	name  string
}

// Run command.
func (hc *HelloCommand) Run(ctx context.Context, args ...string) error {
	if hc.State.Verbose {
		fmt.Println("Starting command...")
	}
	fmt.Printf("Hello, %s!\n", hc.name)
	return nil
}

// Name of the command.
func (hc *HelloCommand) Name() string {
	return "hello"
}

// Short description of the command.
func (hc *HelloCommand) Short() string {
	return `say hello!`
}

// Flags of the command.
func (hc *HelloCommand) Flags(flags *flag.FlagSet) {
	flags.StringVar(&hc.name, "name", "World", "your name")
}

func main() {
	state := &State{}
	p := clino.Program{
		Root: &RootCommand{
			State: state,
		},
		GlobalFlags: state.Flags,
	}
	if err := p.Run(context.Background(), os.Args[1:]...); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(clino.ExitCode(err))
	}
}
