package cmd_test

import (
	"fmt"

	"github.com/posener/cmd"
)

// In the cmd package, positional arguments should be explicitly defined. They are defined using the
// `Args` or `ArgsVar` methods.
func Example_args() {
	var (
		root = cmd.New()
		// Positional arguments should be defined as any other flag.
		args = root.Args("[args...]", "positional arguments for command line")
	)

	// Parse fake command line arguments.
	root.Parse([]string{"cmd", "v1", "v2", "v3"})

	// Test:

	fmt.Println(*args)
	// Output: [v1 v2 v3]
}

// An example of defining an exact number of positional arguments.
func Example_argsN() {
	// Should be defined in global `var`.
	var (
		root = cmd.New()
		// Define a variable that will hold positional arguments. Create the `ArgsStr` object with
		// cap=2 to ensure that the number of arguments is exactly 2.
		args = make(cmd.ArgsStr, 2)
	)

	// Should be in `init()`.
	// Register the positional argument variable in the root command using the `ArgsVar` method
	// (similar to the Var methods of the standard library).
	root.ArgsVar(&args, "[src] [dst]", "positional arguments for command line")

	// Should be in `main()`.
	// Parse fake command line arguments.
	root.Parse([]string{"cmd", "from.txt", "to.txt"})

	// Test:

	fmt.Println(args)
	// Output: [from.txt to.txt]
}

// An example of defining int positional arguments.
func Example_argsInt() {
	// Should be defined in global `var`.
	var (
		root = cmd.New()
		// Define a variable that will hold the positional arguments values. Use the `ArgsInt` type
		// to parse them as int.
		args cmd.ArgsInt
	)

	// Should be in `init()`.
	// Register the positional argument variable in the root command using the `ArgsVar` method.
	root.ArgsVar(&args, "[int...]", "numbers to sum")

	// Should be in `main()`.
	// Parse fake command line arguments.
	root.Parse([]string{"cmd", "10", "20", "30"})

	// Test:

	sum := 0
	for _, n := range args {
		sum += n
	}
	fmt.Println(sum)
	// Output: 60
}

// An example of how to parse positional arguments using a custom function. It enables the advantage
// of using named variables such as `src` and `dst` as opposed to args[0] and args[1].
func Example_argsFn() {
	// Should be defined in global `var`.
	var (
		root = cmd.New()
		// Define variables that will hold the command line positional arguments.
		src, dst string
	)

	// Define an `ArgsFn` that converts a list of positional arguments to the named variables. It
	// should return an error when the arguments are invalid.
	argsFn := cmd.ArgsFn(func(args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("expected src and dst, got %d arguments", len(args))
		}
		src, dst = args[0], args[1]
		return nil
	})

	// Should be in `init()`.
	// Register the function in the root command using the `ArgsVar` method.
	root.ArgsVar(argsFn, "[src] [dst]", "positional arguments for command line")

	// Should be in `main()`.
	root.Parse([]string{"cmd", "from.txt", "to.txt"})

	// Test:

	fmt.Println(src, dst)
	// Output: from.txt to.txt
}
