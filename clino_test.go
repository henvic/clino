package clino

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io/ioutil"
	"os/exec"
	"reflect"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestProgramNoArguments(t *testing.T) {
	var buf bytes.Buffer
	sc := &simpleCommand{}
	p := Program{
		Root:   sc,
		Output: &buf,
	}
	if err := p.Run(context.Background()); err != nil {
		t.Errorf("wanted error to be nil, got %v instead", err)
	}

	if buf.Len() != 0 {
		t.Errorf("got unexpected output, should be empty: %v", buf.String())
	}

	if !sc.ran {
		t.Error("run function wasn't called")
	}

	if len(sc.args) != 0 {
		t.Errorf("expected no arguments to be passed, got %v instead", sc.args)
	}
}

func TestProgramNoArgumentsDoubleHyphenMinus(t *testing.T) {
	var buf bytes.Buffer
	sc := &simpleCommand{}
	p := Program{
		Root:   sc,
		Output: &buf,
	}
	// -- should be ignored and not passed up to the command.
	if err := p.Run(context.Background(), "--"); err != nil {
		t.Errorf("wanted error to be nil, got %v instead", err)
	}

	if buf.Len() != 0 {
		t.Errorf("got unexpected output, should be empty: %v", buf.String())
	}

	if !sc.ran {
		t.Error("run function wasn't called")
	}

	if len(sc.args) != 0 {
		t.Errorf("expected no arguments to be passed, got %v instead", sc.args)
	}
}

func TestProgramArguments(t *testing.T) {
	var buf bytes.Buffer
	sc := &simpleCommand{}
	p := Program{
		Root:   sc,
		Output: &buf,
	}
	// passing "dangerous" arguments (like, unparsed flags after "--" - which should be not parsed by the flag package)
	if err := p.Run(context.Background(), "abc", "def", "123", "--", "help", "xyz", "-name", "Gopher", "-h"); err != nil {
		t.Errorf("wanted error to be nil, got %v instead", err)
	}

	if buf.Len() != 0 {
		t.Errorf("got unexpected output, should be empty: %v", buf.String())
	}

	if !sc.ran {
		t.Error("run function wasn't called")
	}

	wantArgs := []string{"abc", "def", "123", "--", "help", "xyz", "-name", "Gopher", "-h"}
	if n := len(wantArgs); len(sc.args) != n {
		t.Errorf("expected %d arguments to be passed, got %v instead", n, sc.args)
	}

	if !reflect.DeepEqual(wantArgs, sc.args) {
		t.Errorf("expected arguments %v, got %v instead", wantArgs, sc.args)
	}
}

func TestProgramGopher(t *testing.T) {
	var buf bytes.Buffer
	sc := &simpleCommand{}
	p := Program{
		Root:   sc,
		Output: &buf,
	}
	if err := p.Run(context.Background(), "-name", "Gopher"); err != nil {
		t.Errorf("wanted error to be nil, got %v instead", err)
	}

	if buf.Len() != 0 {
		t.Errorf("got unexpected output, should be empty: %v", buf.String())
	}

	if !sc.ran {
		t.Error("run function wasn't called")
	}

	if n := 0; len(sc.args) != n {
		t.Errorf("expected %d arguments to be passed, got %v instead", n, sc.args)
	}

	if want := "Gopher"; sc.name != want {
		t.Errorf("expected name flag to be %q, got %v instead", want, sc.name)
	}
}

func TestProgramHelp(t *testing.T) {
	testCases := []struct {
		desc    string
		program Program
		args    []string
		err     string
		golden  string // be careful: golden files are reused below.
	}{
		{
			desc:    "simple program called with -h",
			program: Program{Root: &simpleCommand{}},
			args:    []string{"-h"},
			golden:  "testdata/simple_help.golden",
		},
		{
			desc: "simple program called with -h and global flags",
			program: Program{
				Root: &simpleCommand{},
				GlobalFlags: func(flags *flag.FlagSet) {
					var globalflag string
					flags.StringVar(&globalflag, "globalflag", "none", "global flag")
				},
			},
			args:   []string{"-h"},
			golden: "testdata/simple_help_global_flags.golden",
		},
		{
			desc:    "simple program called with -help",
			program: Program{Root: &simpleCommand{}},
			args:    []string{"-help"},
			golden:  "testdata/simple_help.golden",
		},
		{
			desc:    "simple program called with --help",
			program: Program{Root: &simpleCommand{}},
			args:    []string{"--help"},
			golden:  "testdata/simple_help.golden",
		},
		{
			desc:    "simple program called with help",
			program: Program{Root: &simpleCommand{}},
			args:    []string{"help"},
			golden:  "testdata/simple_help.golden",
		},
		{
			desc:    "simple program called with undefined flag",
			program: Program{Root: &simpleCommand{}},
			args:    []string{"-undefined"},
			err:     "flag provided but not defined: -undefined",
		},
		{
			desc:    "multiple commands program",
			program: Program{Root: &rootCommand{}},
			golden:  "testdata/root_help.golden",
		},
		{
			desc:    "multiple commands program called with help",
			program: Program{Root: &rootCommand{}},
			args:    []string{"help"},
			golden:  "testdata/root_help.golden",
		},
		{
			desc:    "multiple commands program called with unimplemented command",
			program: Program{Root: &rootCommand{}},
			args:    []string{"unimplemented"},
			err:     "command or topic 'unimplemented' is missing implementation",
		},
		{
			desc:    "multiple commands program called with help -h",
			program: Program{Root: &rootCommand{}},
			args:    []string{"help", "-h"},
			golden:  "testdata/root_help.golden",
		},
		{
			desc:    "multiple commands program called with help -h command-not-found",
			program: Program{Root: &rootCommand{}},
			args:    []string{"help", "command-not-found"},
			err:     "unknown command: 'app command-not-found'",
			golden:  "testdata/root_help.golden",
		},
		{
			desc:    "not-runnable command called with help",
			program: Program{Root: &rootCommand{}},
			args:    []string{"help", "not-runnable"},
			golden:  "testdata/root_not_runnable_help.golden",
		},
		{
			desc:    "not-runnable command called with -h",
			program: Program{Root: &rootCommand{}},
			args:    []string{"not-runnable", "-h"},
			golden:  "testdata/root_not_runnable_help.golden",
		},
		{
			desc:    "--help not-runnable command called",
			program: Program{Root: &rootCommand{}},
			args:    []string{"help", "not-runnable"},
			golden:  "testdata/root_not_runnable_help.golden",
		},
		{
			desc:    "not-runnable command called with -help",
			program: Program{Root: &rootCommand{}},
			args:    []string{"not-runnable", "-help"},
			golden:  "testdata/root_not_runnable_help.golden",
		},
		{
			desc:    "not-runnable command called with --help",
			program: Program{Root: &rootCommand{}},
			args:    []string{"not-runnable", "--help"},
			golden:  "testdata/root_not_runnable_help.golden",
		},
		{
			desc:    "command with flags",
			program: Program{Root: &rootCommandWithFlags{}},
			golden:  "testdata/commands_with_flags.golden",
			args:    []string{"-h"},
		},
		{
			desc:    "inner command with children -h",
			program: Program{Root: &rootCommandWithFlags{}},
			golden:  "testdata/inner_commands_with_children.golden",
			args:    []string{"help", "inner"},
		},
		{
			desc:    "-help inner command with children",
			program: Program{Root: &rootCommandWithFlags{}},
			golden:  "testdata/inner_commands_with_children.golden",
			args:    []string{"inner", "-help"},
		},
		{
			desc:    "inner command with children -help",
			program: Program{Root: &rootCommandWithFlags{}},
			golden:  "testdata/inner_commands_with_children.golden",
			args:    []string{"inner", "-help"},
		},
		{
			desc: "inner command with children -help and global flags",
			program: Program{
				Root: &rootCommandWithFlags{},
				GlobalFlags: func(flags *flag.FlagSet) {
					var globalflag string
					flags.StringVar(&globalflag, "globalflag", "none", "global flag")
				},
			},
			golden: "testdata/inner_commands_with_children_global_flags.golden",
			args:   []string{"inner", "-help"},
		},
		{
			desc:    "not runnable inner subcommand",
			program: Program{Root: &rootCommandWithFlags{}},
			golden:  "testdata/inner_not_runnable_command.golden",
			args:    []string{"inner", "not-runnable"},
		},
		{
			desc:    "inner simple subcommand",
			program: Program{Root: &rootCommandWithFlags{}},
			err:     "flag provided but not defined: -undefined",
			args:    []string{"inner", "simple", "-undefined"},
		},
		{
			desc:    "ignored command not found",
			program: Program{Root: &rootCommandWithFlags{}},
			args:    []string{"help", "ignored-notfound"},
			golden:  "testdata/ignored_notfound_command.golden",
			// 'unknown command' error message is not printed because the rootCommandWithFlags command is runnable,
			// therefore this can be interpreted as its arguments.
		},
		{
			desc:    "command not found",
			program: Program{Root: &rootCommand{}},
			args:    []string{"notfound", "x", "-v"},
			err:     "unknown command: 'app notfound'",
			golden:  "testdata/help_notfound_command.golden",
		},
		{
			desc:    "help command not found",
			program: Program{Root: &rootCommandWithFlags{}},
			args:    []string{"help", "inner", "notfound", "x", "-v"},
			err:     "unknown command: 'cmd inner notfound'",
			golden:  "testdata/inner_notfound_command.golden",
		},
		{
			desc: "help command not found with global flags",
			program: Program{
				Root: &rootCommandWithFlags{},
				GlobalFlags: func(flags *flag.FlagSet) {
					var globalflag string
					flags.StringVar(&globalflag, "globalflag", "none", "global flag")
				},
			},
			args:   []string{"help", "inner", "notfound", "x", "-v"},
			err:    "unknown command: 'cmd inner notfound'",
			golden: "testdata/inner_notfound_command_global_flags.golden",
		},
		{
			desc:    "inner command not found -help",
			program: Program{Root: &rootCommandWithFlags{}},
			args:    []string{"inner", "notfound", "-help", "x", "-v"},
			err:     "unknown command: 'cmd inner notfound'",
			golden:  "testdata/inner_notfound_command.golden",
		},
		{
			desc: "inner command not found -help with global flags",
			program: Program{
				Root: &rootCommandWithFlags{},
				GlobalFlags: func(flags *flag.FlagSet) {
					var globalflag string
					flags.StringVar(&globalflag, "globalflag", "none", "global flag")
				},
			},
			args:   []string{"inner", "notfound", "-help", "x", "-v"},
			err:    "unknown command: 'cmd inner notfound'",
			golden: "testdata/inner_notfound_command_global_flags.golden",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			var buf bytes.Buffer
			tc.program.Output = &buf
			err := tc.program.Run(context.Background(), tc.args...)
			if (err == nil && tc.err != "") || (err != nil && err.Error() != tc.err) {
				t.Errorf("wanted error to be %v, got %v instead", tc.err, err)
			}

			got := buf.String()
			if tc.golden == "" {
				if buf.Len() == 0 {
					return
				}
				t.Errorf("got output %v\n, but found no golden file", got)
			}
			if *update {
				if err = ioutil.WriteFile(tc.golden, buf.Bytes(), 0666); err != nil {
					t.Fatal(err)
				}
			}
			bs, err := ioutil.ReadFile(tc.golden)
			if err != nil {
				t.Fatalf("opening %s: %v", tc.golden, err)
			}
			if got != string(bs) {
				t.Errorf("got output %v\n, wanted %v", got, string(bs))
			}
		})
	}
}

func TestRunRootCommandNotImplemented(t *testing.T) {
	want := "root command not implemented"
	defer func() {
		if r := recover(); r.(string) != want {
			t.Errorf("expected panic message not found, got %v instead", r)
		}
	}()
	var p Program
	t.Fatal(p.Run(context.Background()))
}

func TestRunCommandImplementedMultipleTimes(t *testing.T) {
	want := "command implemented multiple times: 'bad simple'"
	defer func() {
		if r := recover(); r.(string) != want {
			t.Errorf("expected panic message not found, got %v instead", r)
		}
	}()
	p := Program{
		Root: &badRootCommand{},
	}
	t.Fatal(p.Run(context.Background()))
}

func TestExitCode(t *testing.T) {
	testCases := []struct {
		desc string
		in   error
		want int
	}{
		{
			desc: "no error",
			in:   nil,
			want: 0,
		},
		{
			desc: "exit error 2",
			in: ExitError{
				Code: 2,
				Err:  errors.New("cannot find system"),
			},
			want: 2,
		},
		{
			desc: "exit default error code",
			in:   errors.New("cannot find error code"),
			want: 1,
		},
		{
			desc: "copy regular program error code",
			// explanation: by default, Go binary exits with error code = 2.
			in: func() error {
				cmd := exec.Command("go")
				return cmd.Run()
			}(),
			want: 2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			if got := ExitCode(tc.in); got != tc.want {
				t.Errorf("wanted ExitCode(%v) = %v, got %v instead", tc.in, tc.want, got)
			}
		})
	}
}

func TestExitError(t *testing.T) {
	err := errors.New("this is the original error")
	ee := ExitError{
		Code: 3,
		Err:  err,
	}
	if ee.Unwrap() != err {
		t.Errorf("expected unwrapped error to be %v", err)
	}
	if ee.Error() != err.Error() {
		t.Error("expected wrapped error to print the same error message")
	}
}
