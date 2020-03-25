# clino
[![GoDoc](https://godoc.org/github.com/henvic/clino?status.svg)](https://godoc.org/github.com/henvic/clino) [![Build Status](https://travis-ci.org/henvic/clino.svg?branch=master)](https://travis-ci.org/henvic/clino) [![Coverage Status](https://coveralls.io/repos/henvic/clino/badge.svg)](https://coveralls.io/r/henvic/clino) [![Go Report Card](https://goreportcard.com/badge/github.com/henvic/clino)](https://goreportcard.com/report/github.com/henvic/clino)

Package clino provides a simple way to create CLI (command-line interface) tools.

You can create commands to use with this package by implementing its interfaces. It supports the Unix -flag style.

[![asciicast](https://asciinema.org/a/313448.svg)](https://asciinema.org/a/313448)

## Implementing commands
With clino, you implement a command by fulfilling interfaces instead of initializing command structs. The main advantage of this approach is that it is easier to test and to avoid globals.

Tip: Go interfaces are fulfilled implicit, so if you don't see a command being exposed or failing (for example), one of the first things you want to do is to check if your methods signatures fulfill clino's interfaces.

### Command interface
Name (or usage line) for the command.

```go
type Command interface {
	Name() string
}
```

### Shorter interface
Shorter description of a command to show in the "help" output on a list of commands.

```go
type Shorter interface {
	Short() string
}
```

### Runnable interface
You should implement this interface for any command that you want to run directly on the CLI.

* It should receive a context and the command arguments, after parsing any flags.
* A context is required as we want cancelation to be a first-class citizen.
* You can rely on the context for canceling long tasks during tests.

```go
type Runnable interface {
	Run(ctx context.Context, args ...string) error
}
```

### FlagSet interface
You want to implement this interface to accept flags on your command. See also Program.GlobalFlags for supporting global flags on your application.

```go
type FlagSet interface {
	Flags(flags *flag.FlagSet)
}
```

### Longer interface
Description or help message for your command.
The help command prints the returned value of the Long function as the "help" output of a command.

```go
type Longer interface {
	Long() string
}
```

### Footer interface
Footer of a command shown in the "help <command>" output. It is useful for things like printing examples.

```go
type Footer interface {
	Foot() string
}
```

### Parent interface
Parent contains all subcommands of a given command. Your root command needs to implement it.
```
type Parent interface {
	Commands() []Command
}
```

### Example code
You can see more examples in the example directory.

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/henvic/clino"
)

func main() {
	p := clino.Program{
		Root: &RootCommand{},
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
```
