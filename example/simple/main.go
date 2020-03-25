package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/henvic/clino"
)

func main() {
	rc := &RootCommand{}
	p := clino.Program{
		Root: rc,
	}
	if err := p.Run(context.Background(), os.Args[1:]...); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(clino.ExitCode(err))
	}
}

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
