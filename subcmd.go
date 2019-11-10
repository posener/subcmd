// subcmd is a minimalistic library that enables easy sub commands with the standard `flag` library.
//
// Define a root command object using the `New` function.
// This object exposes the standard library's `flag.FlagSet` API, which enables adding flags in the
// standard way.
// Additionally, this object exposes the `SubCommand` method, which returns another command object.
// This objects also exposing the same API, enabling definition of flags and nested sub commands.
//
// The root object then have to be called with the `Parse` or `ParseArgs` methods, similarly to
// the `flag.Parse` call.
//
// The usage is automatically configured to show both sub commands and flags.
//
// Automatic bash completion is enabled for command sub commands and flag names.
//
// Principles
//
// * Minimalistic and `flag`-like.
//
// * Any flag that is defined in the base command will be reflected in all of its sub commands.
//
// * When user types the command, it starts from the command and sub commands, only then types the
// flags and then the positional arguments:
//
// 	[command] [sub commands...] [flags...] [positional args...]
//
// * Positional arguments are as any other flag: their number and type should be enforced and
// checked.
//
// * When a command that defined positional arguments, all its sub commands has these positional
// arguments and thus can't define their own positional arguments.
//
// * Usage format is standard, programs can't define their own format.
//
// * When flag configuration is wrong, the program will panic when starts. Most of them in flag
// definition stage, and not after flag parsing stage.
package subcmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/posener/complete/v2"
	"github.com/posener/formatter"
)

// Cmd is a command that can have set of flags and sub commands.
type Cmd struct {
	*SubCmd
}

// SubCmd is a sub command that can have a set of flags and sub commands.
type SubCmd struct {
	config
	// flagsSet holds the flags of the command.
	flagSet *flag.FlagSet
	// sub holds the sub commands of the command.
	sub map[string]*SubCmd
	// args are the positional arguments. If nil the command does not accept positional arguments.
	args *argsData

	isRoot bool
}

// argsData contains data about argsData arguments.
type argsData struct {
	value          ArgsValue
	usage, details string
}

// ArgsValue is interface for positional arguments variable. It can be used with the
// `(*Cmd).ArgsVar` method. For examples of objects that implement this interface see ./args.go.
type ArgsValue interface {
	// Set should assign values to the positional arguments variable from list of positional
	// arguments from the command line. It should return an error if the given list does not fit
	// the requirements.
	Set([]string) error
}

// ArgsFn is a function that implements Args. Usage example:
//
// 	var (
// 		cmd      = subcmd.Root()
// 		src, dst string
// 	)
//
// 	func setArgs(args []string) error {
// 		if len(args) != 2 {
// 			return fmt.Errorf("expected src and dst, got %d arguments", len(args))
// 		}
// 		src, dst = args[0], args[1]
// 		return nil
// 	}
//
// 	func init() {
// 		cmd.ArgsVar(subcmd.ArgsFn(setArgs), "[src] [dst]", "define source and destination")
// 	}
type ArgsFn func([]string) error

func (f ArgsFn) Set(args []string) error { return f(args) }

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

// New creates a new root command.
func New(options ...optionRoot) *Cmd {
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

// Parse command line arguments.
func (c *Cmd) ParseArgs() error {
	return c.Parse(os.Args)
}

// Parse a set of arguments.
func (c *Cmd) Parse(args []string) error {
	c.complete(args)
	_, err := c.parse(args)
	return c.handleError(err)
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

// SubCommand creates a new sub command to the given command.
func (c *SubCmd) SubCommand(name string, synopsis string, options ...option) *SubCmd {
	if len(name) == 0 {
		panic("subcommand can't be empty")
	}
	if name[0] == '-' {
		panic("subcommand can't start with a dash")
	}
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

	subCmd := newSubCmd(cfg, c.flagSet)
	subCmd.args = c.args

	c.sub[name] = subCmd
	return subCmd
}

// Args returns the positional arguments for the command and enable defining options. Only a sub
// command that called this method accepts positional arguments. Calling a sub command with
// positional arguments where they were not defined result in parsing error. The provided options
// can be nil for default values.
func (c *SubCmd) Args(usage, details string) *[]string {
	var args ArgsStr
	c.ArgsVar(&args, usage, details)
	return (*[]string)(&args)
}

// ArgsVar should be used to parse arguments with specific requirements or to specific object/s.
// For example, accept only 3 positional arguments:
//
// 	var (
// 		cmd  = subcmd.Root()
// 		args = make(subcmd.ArgsStr, 3)
// 	)
//
// 	func init() {
// 		cmd.ArgsVar(args, "[arg1] [arg2] [arg3]", "provide 3 positional arguments")
// 	}
func (c *SubCmd) ArgsVar(value ArgsValue, usage, details string) {
	c.checkNewArgs()
	if c.args != nil {
		panic("Args() or ArgsVar() called more than once.")
	}
	c.args = &argsData{
		value:   value,
		usage:   usage,
		details: details,
	}

	if c.args.usage == "" {
		c.args.usage = "[args...]"
	}
}

func (c *SubCmd) parse(args []string) ([]string, error) {
	if len(args) < 1 {
		panic("must be at least the command in arguments")
	}

	// First argument is the command name.
	args = args[1:]

	/// If command has sub commands, find it and parse the sub command.
	if len(c.sub) > 0 {
		if len(args) == 0 {
			return nil, fmt.Errorf("must provide sub command")
		}
		cmd := args[0]
		if c.sub[cmd] == nil {
			// Check for help flag, which can be applied on any level of sub command.
			if cmd == "-h" || cmd == "-help" || cmd == "--help" {
				c.Usage()
				return nil, flag.ErrHelp
			}
			return nil, fmt.Errorf("invalid command: %s", cmd)
		}
		var err error
		args, err = c.sub[cmd].parse(args)
		if err != nil {
			return nil, fmt.Errorf("%s > %v", c.name, err)
		}
	}

	// Check for command flags, and update the remaining arguments.
	err := c.flagSet.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("%s: bad flags: %w", c.name, err)
	}
	args = c.flagSet.Args()

	// Collect positional arguments if required.
	args, err = c.setArgs(args)
	if err != nil {
		return nil, fmt.Errorf("%s: bad positional args: %v", c.name, err)
	}

	return args, nil
}

func (c *SubCmd) setArgs(args []string) ([]string, error) {
	if c.args == nil {
		if len(args) > 0 {
			return nil, fmt.Errorf("positional args not expected, got %v", args)
		}
		return args, nil
	}
	return nil, c.args.value.Set(args)
}

func (c *SubCmd) Usage() {
	w := c.output
	detailsW := detailsWriter(w)
	subs := c.subNames()

	// Constract usage string.

	usage := "Usage: " + c.name
	if len(subs) == 0 {
		if c.hasFlags() {
			usage += " [flags]"
		}
		if c.args != nil {
			usage += " " + c.args.usage
		}
	} else {
		subcommands := "[" + strings.Join(subs, "|") + "]"
		if len(subcommands) > 30 {
			subcommands = "[subcommands...]"
		}
		usage += " " + subcommands
	}

	// Add synopsis and details.

	fmt.Fprintf(w, usage+"\n\n")
	if c.synopsis != "" {
		fmt.Fprintf(w, c.synopsis+"\n\n")
	}
	if c.details != "" {
		fmt.Fprintf(detailsW, c.details)
		fmt.Fprintf(w, "\n\n")
	}

	// Describe sub commands or flags and positional arguments.

	if len(c.sub) > 0 {
		fmt.Fprintf(w, "Subcommands:\n\n")
		for _, name := range subs {
			fmt.Fprintf(w, "  %s\t%s\n", name, c.sub[name].synopsis)
		}
		fmt.Fprintf(w, "\n")
		// Print completion options only to the root command.
		if c.isRoot {
			fmt.Fprintln(w, completionUsage(c.name))
		}
	} else {
		if c.hasFlags() {
			fmt.Fprintf(w, "Flags:\n\n")
			c.flagSet.PrintDefaults()
			fmt.Fprintf(w, "\n")
		}

		if c.args != nil && c.args.details != "" {
			fmt.Fprintf(w, "Positional arguments:\n\n")
			fmt.Fprintf(detailsW, c.args.details)
			fmt.Fprintf(w, "\n\n")
		}
	}
}

// subNames return all sub commands ordered alphabetically.
func (c *SubCmd) subNames() []string {
	names := make([]string, 0, len(c.sub))
	for name := range c.sub {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *SubCmd) hasFlags() bool {
	hasFlags := false
	c.flagSet.VisitAll(func(*flag.Flag) { hasFlags = true })
	return hasFlags
}

// complete performs bash completion when required.
func (c *Cmd) complete(args []string) {
	complete.Complete(c.name, (*completer)(c.SubCmd))
}

func newCmd(cfg config) *Cmd {
	cmd := &Cmd{SubCmd: newSubCmd(cfg, nil)}
	cmd.isRoot = true
	return cmd
}

func newSubCmd(cfg config, parentFs *flag.FlagSet) *SubCmd {
	subcmd := &SubCmd{
		config:  cfg,
		flagSet: copyFlagSet(cfg, parentFs),
		sub:     make(map[string]*SubCmd),
	}
	subcmd.flagSet.Usage = subcmd.Usage
	return subcmd
}

func detailsWriter(w io.Writer) io.Writer {
	return &formatter.Formatter{Writer: w, Width: 80, Indent: []byte("  ")}
}

func copyMap(m map[string]string) map[string]string {
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func copyFlagSet(cfg config, f *flag.FlagSet) *flag.FlagSet {
	cp := flag.NewFlagSet(cfg.name, flag.ContinueOnError)
	cp.SetOutput(cfg.output)
	if f != nil {
		f.VisitAll(func(fl *flag.Flag) { cp.Var(fl.Value, fl.Name, fl.Usage) })
	}
	return cp
}

func completionUsage(name string) string {
	return fmt.Sprintf(`Bash Completion:

Install bash completion by running: 'COMP_INSTALL=1 %s'.
Uninstall by running: 'COMP_UNINSTALL=1 %s'.
Skip installation prompt with environment variable: 'COMP_YES=1'.
`, name, name)
}
