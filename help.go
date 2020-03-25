package clino

import (
	"context"
	"flag"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

// helper is invoked when function is called with no arguments, directly, or with the "help" command or -help flag.
type helper struct {
	Output io.Writer

	Long func() string
	Foot func() string

	Commands []Command

	binary   string
	trail    []string
	args     []string
	runnable bool
	usable   bool

	fs *flag.FlagSet
}

func argumentsNonFlags(args []string) (nargs []string) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			return
		}
		nargs = append(nargs, arg)
	}
	return
}

// Run help command.
func (h *helper) Run(ctx context.Context) (err error) {
	defer func() {
		na := argumentsNonFlags(h.args)
		if err == nil && !h.runnable && len(na) > len(h.trail) {
			err = commandNotFound(h.binary, na[:len(h.trail)+1])
		}
	}()
	if h.Long != nil {
		fmt.Fprintf(h.Output, "%s\n", h.Long())
	}
	var xcmd string
	command := strings.Join(h.trail, " ")
	if command == "" {
		command = "<command>"
	} else if len(h.Commands) != 0 {
		xcmd = " <command>"
	}
	if h.Long != nil && h.usable {
		fmt.Fprintln(h.Output)
	}
	if h.usable {
		fmt.Fprintf(h.Output, "Usage:  %s %s%s [flags] [arguments]\n\n", h.binary, command, xcmd)
	}
	w := tabwriter.NewWriter(h.Output, 0, 0, 8, ' ', 0)
	h.helpCommands(w)
	if h.usable {
		h.helpFlags(w)
	}
	if len(h.Commands) != 0 {
		fmt.Fprintf(w, "Use \"%s help %s%s\" for more information about that command.\n", h.binary, command, xcmd)
	}
	if err = w.Flush(); err != nil {
		return err
	}
	if h.Foot != nil {
		fmt.Fprintf(h.Output, "%s\n", h.Foot())
	}
	if !h.usable && h.Long == nil && h.Foot == nil {
		// useful commands should implement at least one of the following interfaces:
		// Runnable, Longer, Parent, or Footer interfaces.
		return fmt.Errorf("command or topic '%v' is missing implementation", strings.Join(h.trail, " "))
	}
	return nil
}

func (h *helper) helpCommands(w io.Writer) {
	if len(h.Commands) == 0 {
		return
	}
	fmt.Fprint(w, "\tCommands:\n\t")
	for _, c := range h.Commands {
		var short string
		if s, ok := c.(Shorter); ok {
			short = s.Short()
		}
		fmt.Fprintf(w, "%s\t%s\n\t", c.Name(), short)
	}
	fmt.Fprintln(w, "\t\t")
}

func (h *helper) helpFlags(w io.Writer) {
	fmt.Fprintln(w, "\tFlags:\t") // \t\t keeps the alignment between commands and flags on tabwriter
	if h.fs != nil {
		h.fs.VisitAll(func(f *flag.Flag) {
			printFlag(w, f)
		})
	}
	fmt.Fprint(w, "\t-help\tshow help message\n\n")
}

func printFlag(w io.Writer, f *flag.Flag) {
	typ, usage := flag.UnquoteUsage(f)
	if typ == "" { // type: bool flag
		fmt.Fprintf(w, "\t-%s\t%s", f.Name, usage)
	} else {
		fmt.Fprintf(w, "\t-%s (%s)\t%s", f.Name, typ, usage)
	}
	if isZeroValue(f, f.DefValue) {
		fmt.Fprintln(w)
		return
	}
	if typ == "string" {
		fmt.Fprintf(w, " (default %q)\n", f.DefValue) // put quotes on the value
		return
	}
	fmt.Fprintf(w, " (default %v)\n", f.DefValue)
}

// isZeroValue determines whether the string represents the zero
// value for a flag.
// Function copied from the Go standard library flag package.
// Source: https://github.com/golang/go/blob/5b15941c61f478b8ed08b76a27186527ba73d273/src/flag/flag.go#L447
func isZeroValue(f *flag.Flag, value string) bool {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(f.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Ptr {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	return value == z.Interface().(flag.Value).String()
}
