// subcmd is a minimalistic library that enables easy sub commands with the standard `flag` library.
//
// Define a `root` command object using the `Root` function.
// This object exposes the standard library's `flag.FlagSet` API, which enables adding flags in the
// standard way.
// Additionally, this object exposes the `SubCommand` method, which returns another command object.
// This objects also exposing the same API, enabling definition of flags and nested sub commands.
//
// The root object then have to be called with the `Parse` or `ParseArgs` methods, similiraly to
// the `flag.Parse` call.
//
// The usage is automatically configured to show both sub commands and flags.
//
// Positional arguments
//
// The `subcmd` library is opinionated about positional arguments: it enforces their definition
// and parsing. The user can define for each sub command if and how many positional arguments it
// accepts. Their usage is similar to the flag values usage.
//
// Example
//
// See ./example/main.go.
//
// Limitations
//
// Suppose `cmd` has a flag `-flag`, and a subcommand `sub`. In the current implementation:
// Calling `cmd sub -flag` won't work as the flag is set after the sub command, while
// `cmd -flag sub` will work perfectly fine. Each flag needs to be used in the scope of its command.
package subcmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/posener/subcmd/internal/formatter"
)

// Cmd is a command that can have set of flags and a sub command.
type Cmd struct {
	config
	// argsOpts are the positional arguments options. If nil the command does not accept
	// positional arguments.
	argsOpts *ArgsOpts

	// FlagsSet holds the flags of the command.
	*flagSet
	// args holds the positional arguments of the commands.
	args *[]string
	// sub holds the sub commands of the command.
	sub map[string]*Cmd
	// output is used to write the usage output.
}

// ArgsOpts are options for positional arguments.
type ArgsOpts struct {
	// N can be used to enforce a fixed number of positional arguments. Any non-positive number will
	// be ignored.
	N int
	// Usage is a string representing the positional arguments which will be printed in the command
	// usage. For example, The string "[source] [destination]" can represent usage of positional
	// arguments for a move command.
	Usage string
	// Details can be used to provide farther explaination on the positional arguments in the usage
	// help.
	Details string
}

// config is configuration for root command.
type config struct {
	subConfig
	name          string
	errorHandling flag.ErrorHandling
	output        io.Writer
}

// subConfig is configuration that used both for root command and sub commands.
type subConfig struct {
	synopsis string
	details  string
}

// optionRoot is an option that can be applied only on the root command and not on sub commands.
type optionRoot interface {
	applyRoot(o *config)
}

// option is an option for configuring a sub commands.
type option interface {
	apply(o *subConfig)
}

// optionRootFn is an option function that can be applied only on the root command and not on sub
// commands.
type optionRootFn func(cfg *config)

func (f optionRootFn) applyRoot(cfg *config) { f(cfg) }

// optionFn is an option function that can be applied on a root command or sub commands.
type optionFn func(cfg *subConfig)

func (f optionFn) applyRoot(cfg *config) { f(&cfg.subConfig) }

func (f optionFn) apply(cfg *subConfig) { f(cfg) }

// OptErrorHandling defines the behavior in case of an error in the `Parse` function.
func OptErrorHandling(errorHandling flag.ErrorHandling) optionRootFn {
	return func(cfg *config) {
		cfg.errorHandling = errorHandling
	}
}

// OptOutput sets the output for the usage.
func OptOutput(w io.Writer) optionRootFn {
	return func(cfg *config) {
		cfg.output = w
	}
}

// OptName sets a predefined name to the root command.
func OptName(name string) optionRootFn {
	return func(cfg *config) {
		cfg.name = name
	}
}

// OptSynopsis sets a description to the root command.
func OptSynopsis(synopsis string) optionRootFn {
	return func(cfg *config) {
		cfg.synopsis = synopsis
	}
}

// OptSynopsis sets a description to the root command.
func OptDetails(details string) optionFn {
	return func(cfg *subConfig) {
		cfg.details = details
	}
}

// Root creats a new root command.
func Root(options ...optionRoot) *Cmd {
	// Set default config.
	cfg := config{
		name:          os.Args[0],
		errorHandling: flag.ExitOnError,
		output:        os.Stderr,
	}
	// Update with requested options.
	for _, option := range options {
		option.applyRoot(&cfg)
	}

	return newCmd(cfg)
}

// SubCommand creates a new sub command to the given command.
func (c *Cmd) SubCommand(name string, synopsis string, options ...option) *Cmd {
	if c.sub[name] != nil {
		panic(fmt.Sprintf("sub command %q already exists", name))
	}

	cfg := c.config
	cfg.name = c.name + " " + name
	cfg.synopsis = synopsis
	cfg.details = ""
	// Update with requested options.
	for _, option := range options {
		option.apply(&cfg.subConfig)
	}

	subCmd := newCmd(cfg)

	c.sub[name] = subCmd
	return subCmd
}

// Args returns the positional arguments for the command and enable defining options. Only a sub
// command that called this method accepts positional arguments. Calling a sub command with
// positional arguments where they were not defined result in parsing error. The provided options
// can be nil for default values.
func (c *Cmd) Args(opts *ArgsOpts) *[]string {
	if c.argsOpts != nil {
		panic("Args() called more than once.")
	}
	// Default options.
	if opts == nil {
		opts = &ArgsOpts{}
	}
	c.argsOpts = opts
	return c.args
}

// Parse command line arguments.
func (c *Cmd) ParseArgs() error {
	return c.Parse(os.Args)
}

// Parse a set of arguments.
func (c *Cmd) Parse(args []string) error {
	c.validate()
	_, err := c.parse(args)

	return c.handleError(err)
}

func (c *Cmd) parse(args []string) ([]string, error) {
	if len(args) < 1 {
		panic("must be at least the command in arguments")
	}

	// Check for command flags, and update the remaining arguments.
	err := c.FlagSet.Parse(args[1:])
	if err != nil {
		return nil, fmt.Errorf("%s: bad flags: %w", c.name, err)
	}
	args = c.FlagSet.Args()

	// Check if another the first remaining argument matches any sub command.
	if len(args) > 0 && c.sub[args[0]] != nil {
		subcmd := c.sub[args[0]]
		args, err = subcmd.parse(args)
		if err != nil {
			return nil, fmt.Errorf("%s > %v", c.name, err)
		}
	}

	// Collect positional arguments if required.
	args, err = c.setArgs(args)
	if err != nil {
		return nil, fmt.Errorf("%s: bad positional args: %v", c.name, err)
	}

	return args, nil
}

func (c *Cmd) setArgs(args []string) ([]string, error) {
	opt := c.argsOpts
	if opt == nil {
		if len(args) > 0 {
			return nil, fmt.Errorf("positional args not expected, got %v", args)
		}
		return args, nil
	}
	if opt.N > 0 && len(args) != opt.N {
		return nil, fmt.Errorf("required %d positional args, got %v", opt.N, args)
	}
	c.args = &args
	return nil, nil
}

// validate the command line. Panics on error.
func (c *Cmd) validate() {
	c.validatePositional("")
}

// validatePositional validates positional arguments. If c was defined with positional arguments,
// any of its sub commands can't be defined with positional arguments.
func (c *Cmd) validatePositional(parentWithPositional string) {
	if hasPositional := c.argsOpts != nil; hasPositional {
		if parentWithPositional != "" {
			panic(fmt.Sprintf(
				"Illegal: parent %q and sub command %q both define positional areguments",
				parentWithPositional, c.name))
		} else {
			parentWithPositional = c.name
		}
	}

	// Check all sub commands.
	for _, subcmd := range c.sub {
		subcmd.validatePositional(parentWithPositional)
	}
}

func (c *Cmd) usage() {
	w := c.output
	detailsW := detailsWriter(w)

	usage := "Usage: " + c.name
	if c.hasFlags() {
		usage += " [flags]"
	}
	if c.argsOpts != nil {
		usage += " " + c.argsOpts.usage()
	}
	fmt.Fprintf(w, usage+"\n\n")
	if c.synopsis != "" {
		fmt.Fprintf(w, c.synopsis+"\n\n")
	}
	if c.details != "" {
		fmt.Fprintf(detailsW, c.details)
		fmt.Fprintf(w, "\n\n")
	}
	if len(c.sub) > 0 {
		fmt.Fprintf(w, "Subcommands:\n\n")
		for _, name := range c.subNames() {
			fmt.Fprintf(w, "  %s\t%s\n", name, c.sub[name].synopsis)
		}
		fmt.Fprintf(w, "\n")
	}

	if c.hasFlags() {
		fmt.Fprintf(w, "Flags:\n\n")
		c.FlagSet.PrintDefaults()
		fmt.Fprintf(w, "\n")
	}

	if c.argsOpts != nil && c.argsOpts.Details != "" {
		fmt.Fprintf(w, "Positional arguments:\n\n")
		fmt.Fprintf(detailsW, c.argsOpts.Details)
		fmt.Fprintf(w, "\n\n")
	}
}

// subNames return all sub commands oredered alphabetically.
func (c *Cmd) subNames() []string {
	names := make([]string, 0, len(c.sub))
	for name := range c.sub {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *Cmd) hasFlags() bool {
	hasFlags := false
	c.FlagSet.VisitAll(func(*flag.Flag) { hasFlags = true })
	return hasFlags
}

func (c *Cmd) handleError(err error) error {
	if err == nil {
		return nil
	}
	switch c.errorHandling {
	case flag.ExitOnError:
		os.Exit(2)
	case flag.PanicOnError:
		panic(err)
	}
	return err
}

func (o *ArgsOpts) usage() string {
	if o.Usage != "" {
		return o.Usage
	}
	return "[args]"
}

func newCmd(cfg config) *Cmd {
	fs := flag.NewFlagSet(os.Args[0], cfg.errorHandling)
	fs.SetOutput(cfg.output)

	cmd := &Cmd{
		config:  cfg,
		flagSet: &flagSet{FlagSet: fs},
		sub:     make(map[string]*Cmd),
	}
	cmd.Usage = cmd.usage
	return cmd
}

func detailsWriter(w io.Writer) io.Writer {
	return &formatter.Formatter{Writer: w, Width: 80, Indent: []byte("  ")}
}
