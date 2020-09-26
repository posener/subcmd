package main

import (
	"fmt"

	"github.com/posener/cmd"
)

var (
	// Define root command with a single string flag. This object the familiar standard library
	// `*flag.FlagSet` API, so it can be used similarly.
	root  = cmd.New()
	flag0 = root.String("flag0", "", "root string flag")

	// Define a sub command from the root command with a single string flag. The sub command object
	// also have the same API as the root command object.
	sub1  = root.SubCommand("sub1", "first sub command")
	flag1 = sub1.String("flag1", "", "sub1 string flag")

	// Define a second sub command from the root command with an int flag.
	sub2  = root.SubCommand("sub2", "second sub command")
	flag2 = sub1.Int("flag2", 0, "sub2 int flag")
)

func main() {
	root.Parse()

	// Check which sub command was chosen by the user.
	switch {
	case sub1.Parsed():
		fmt.Printf("Called sub1 with flag: %s", *flag1)
	case sub2.Parsed():
		fmt.Printf("Called sub2 with flag: %d", *flag2)
	}
}
