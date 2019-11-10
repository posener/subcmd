package subcmd

import (
	"bytes"
	"errors"
	"flag"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/posener/complete/v2"
	"github.com/stretchr/testify/assert"
)

type testCommand struct {
	*Cmd
	rootFlag *bool

	sub1     *SubCmd
	sub1Flag *string

	sub11     *SubCmd
	sub11Flag *string
	sub11Args *[]string

	sub12     *SubCmd
	sub12Flag *string

	sub2     *SubCmd
	sub2Args ArgsStr

	out bytes.Buffer
}

const longText = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

func testNew() *testCommand {
	var cmd testCommand

	cmd.Cmd = New(
		OptName("cmd"),
		OptErrorHandling(flag.ContinueOnError),
		OptOutput(&cmd.out),
		OptSynopsis("cmd synopsis"),
		OptDetails("testing command line example"))

	cmd.rootFlag = cmd.Bool("flag0", false, "example of bool flag")

	cmd.sub1 = cmd.SubCommand("sub1", "a sub command with flags and sub commands", OptDetails(longText))
	cmd.sub1Flag = cmd.sub1.String("flag1", "", "example of string flag")

	cmd.sub11 = cmd.sub1.SubCommand("sub1", "sub command of sub command")
	cmd.sub11Flag = cmd.sub11.String("flag11", "", "example of string flag")
	cmd.sub11Args = cmd.sub11.Args("", "")

	cmd.sub12 = cmd.sub1.SubCommand("sub2", "sub command of sub command")
	cmd.sub12Flag = cmd.sub12.String("flag12", "", "example of string flag")

	cmd.sub2 = cmd.SubCommand("sub2", "a sub command without flags and sub commands")
	cmd.sub2Args = make(ArgsStr, 0, 1)
	cmd.sub2.ArgsVar(&cmd.sub2Args, "[arg]", "arg is a single argument")

	return &cmd
}

func TestSubCmd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		sub1Parsed  bool
		sub11Parsed bool
		sub2Parsed  bool
		rootFlag    bool
		sub1Flag    string
		sub11Flag   string
		sub11Args   []string
		sub2Args    []string
	}{
		{
			name:    "cmd: can't be called without a sub command",
			args:    []string{"cmd"},
			wantErr: true,
		},
		{
			name: "cmd: can be called with help",
			args: []string{"cmd", "-h"},
		},
		{
			name:    "cmd sub1: can't be called without a sub command",
			args:    []string{"cmd", "sub1", "-flag0", "-flag1", "value"},
			wantErr: true,
		},
		{
			name:        "cmd sub1 sub1: with flags and args",
			args:        []string{"cmd", "sub1", "sub1", "-flag11", "value11", "-flag0", "-flag1", "value1", "arg1", "arg2"},
			sub1Parsed:  true,
			sub11Parsed: true,
			rootFlag:    true,
			sub1Flag:    "value1",
			sub11Flag:   "value11",
			sub11Args:   []string{"arg1", "arg2"},
		},
		{
			name:    "cmd sub1 sub2: pass positional argument to a command that does not define positional arguments",
			args:    []string{"cmd", "sub1", "sub2", "arg1"},
			wantErr: true,
		},
		{
			name:       "cmd sub2: with 1 positional arguments",
			args:       []string{"cmd", "sub2", "arg1"},
			sub2Parsed: true,
		},
		{
			name:    "cmd sub2: fails with 2 positional arguments",
			args:    []string{"cmd", "sub2", "arg1", "arg2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := testNew()
			err := cmd.Parse(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.True(t, err == nil || errors.As(err, &flag.ErrHelp))
				assert.Equal(t, tt.sub1Parsed, cmd.sub1.Parsed())
				assert.Equal(t, tt.sub11Parsed, cmd.sub11.Parsed())
				assert.Equal(t, tt.sub2Parsed, cmd.sub2.Parsed())
				assert.Equal(t, tt.rootFlag, *cmd.rootFlag)
				assert.Equal(t, tt.sub1Flag, *cmd.sub1Flag)
				assert.Equal(t, tt.sub11Flag, *cmd.sub11Flag)
			}
		})
	}
}

func TestHelp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args []string
		want string
	}{
		{
			args: []string{"cmd", "-h"},
			want: `Usage: cmd [sub1|sub2]

cmd synopsis

  testing command line example

Subcommands:

  sub1	a sub command with flags and sub commands
  sub2	a sub command without flags and sub commands

Bash Completion:

Install bash completion by running: 'COMP_INSTALL=1 cmd'.
Uninstall by running: 'COMP_UNINSTALL=1 cmd'.
Skip installation prompt with environment variable: 'COMP_YES=1'.

`,
		},
		{
			args: []string{"cmd", "sub1", "-h"},
			want: `Usage: cmd sub1 [sub1|sub2]

a sub command with flags and sub commands

  Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
  incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis
  nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
  Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore
  eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt
  in culpa qui officia deserunt mollit anim id est laborum.

Subcommands:

  sub1	sub command of sub command
  sub2	sub command of sub command

`,
		},
		{
			args: []string{"cmd", "sub2", "-h"},
			want: `Usage: cmd sub2 [flags] [arg]

a sub command without flags and sub commands

Flags:

  -flag0
    	example of bool flag

Positional arguments:

  arg is a single argument

`,
		},
		{
			args: []string{"cmd", "sub1", "sub1", "-h"},
			want: `Usage: cmd sub1 sub1 [flags] [args...]

sub command of sub command

Flags:

  -flag0
    	example of bool flag
  -flag1 string
    	example of string flag
  -flag11 string
    	example of string flag

`,
		},
		{
			args: []string{"cmd", "sub1", "sub2", "-h"},
			want: `Usage: cmd sub1 sub2 [flags]

sub command of sub command

Flags:

  -flag0
    	example of bool flag
  -flag1 string
    	example of string flag
  -flag12 string
    	example of string flag

`,
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			cmd := testNew()
			err := cmd.Parse(tt.args)
			assert.True(t, errors.As(err, &flag.ErrHelp))
			assert.Equal(t, tt.want, cmd.out.String())
		})
	}
}

func TestCmd_failures(t *testing.T) {
	t.Parallel()

	t.Run("subcommand valid names", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))
		assert.Panics(t, func() { cmd.SubCommand("", "") })
		assert.Panics(t, func() { cmd.SubCommand("-name", "") })
	})

	t.Run("command can't have two sub commands with the same name", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))
		cmd.SubCommand("sub", "")

		assert.Panics(t, func() { cmd.SubCommand("sub", "") })
	})

	t.Run("parse must get at least one argument", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))

		assert.Panics(t, func() { cmd.Parse(nil) })
	})

	t.Run("defining flag after subcommand is not allowed", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))
		cmd.SubCommand("sub", "")

		assert.Panics(t, func() { cmd.String("flag", "", "") })
	})

	t.Run("defining args after subcommand is not allowed", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))
		cmd.SubCommand("sub", "")

		assert.Panics(t, func() { cmd.Args("flag", "") })
	})

	t.Run("both command and sub command have the same flag name should panic", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))
		cmd.String("flag", "", "")
		subcmd := cmd.SubCommand("sub", "")

		assert.Panics(t, func() { subcmd.String("flag", "", "") })
	})

	t.Run("both command and sub command have positional arguments should panic", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))
		cmd.Args("", "")
		subcmd := cmd.SubCommand("sub", "")

		assert.Panics(t, func() { subcmd.Args("", "") })
	})

	t.Run("both command and sub sub command have positional arguments should panic", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))
		cmd.Args("", "")
		sub := cmd.SubCommand("sub", "")
		subsub := sub.SubCommand("sub", "")

		assert.Panics(t, func() { subsub.Args("", "") })
	})

	t.Run("both sub command and sub sub command have positional arguments should panic", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))
		sub := cmd.SubCommand("sub", "")
		sub.Args("", "")
		subsub := sub.SubCommand("sub", "")

		assert.Panics(t, func() { subsub.Args("", "") })
	})

	t.Run("two different sub command may have positional arguments", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))
		sub1 := cmd.SubCommand("sub1", "")
		sub1.Args("", "")
		sub2 := cmd.SubCommand("sub2", "")
		sub2.Args("", "")

		assert.NotPanics(t, func() { cmd.Parse([]string{"cmd", "sub1"}) })
	})

	t.Run("calling positional more than once is not allowed", func(t *testing.T) {
		cmd := New(OptOutput(ioutil.Discard))
		cmd.Args("", "")

		assert.Panics(t, func() { cmd.Args("", "") })
	})
}

func TestComplete(t *testing.T) {
	t.Parallel()

	comp := (*completer)(testNew().SubCmd)

	tests := []struct {
		line        string
		completions []string
	}{
		{line: "su", completions: []string{"sub1", "sub2"}},
		{line: "sub1 sub1 -f", completions: []string{"-flag1", "-flag0", "-flag11"}},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			complete.Test(t, comp, tt.line, tt.completions)
		})
	}
}
