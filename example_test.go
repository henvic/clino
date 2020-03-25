package clino_test

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/henvic/clino"
)

// RootCommand is the entrypoint of the application.
type RootCommand struct {
	name string
}

// Name of the application.
func (rc *RootCommand) Name() string {
	return "app"
}

// Long description of the application.
func (rc *RootCommand) Long() string {
	return "Example application."
}

// Flags of the command.
func (rc *RootCommand) Flags(flags *flag.FlagSet) {
	flags.StringVar(&rc.name, "name", "World", "your name")
}

// Run command.
func (rc *RootCommand) Run(ctx context.Context, args ...string) error {
	fmt.Printf("Hello, %s!\n", rc.name)
	return nil
}

func Example() {
	p := clino.Program{
		Root: &RootCommand{},
	}
	if err := p.Run(context.Background(), "-name", "Gopher"); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(clino.ExitCode(err))
	}
	// Output:
	// Hello, Gopher!
}
