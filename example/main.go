package main

import (
	"fmt"

	"github.com/posener/subcmd"
)

var (
	// Define a root command. Some options can be set using the `Opt*` functions. It returns a
	// `*Cmd` object.
	root = subcmd.New()
	// The `*Cmd` object can be used as the standard library `flag.FlagSet`.
	flag0 = root.String("flag0", "", "root stringflag")
	// From each command object, a sub command can be created. This can be done recursively.
	sub1 = root.SubCommand("sub1", "first sub command")
	// Each sub command can have flags attached.
	flag1 = sub1.String("flag1", "", "sub1 string flag")
	sub2  = root.SubCommand("sub2", "second sub command")
	flag2 = sub1.Int("flag2", 0, "sub2 int flag")
	_     = root.SubCommand("xxx", "xxx")
)

// Definition and usage of sub commands and sub commands flags.
func main() {
	root.ParseArgs()
	switch {
	case sub1.Parsed():
		fmt.Printf("Called sub1 with flag: %s", *flag1)
	case sub2.Parsed():
		fmt.Printf("Called sub2 with flag: %d", *flag2)
	}
}
