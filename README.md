# cmd

[![Build Status](https://travis-ci.org/posener/cmd.svg?branch=master)](https://travis-ci.org/posener/cmd)
[![codecov](https://codecov.io/gh/posener/cmd/branch/master/graph/badge.svg)](https://codecov.io/gh/posener/cmd)
[![GoDoc](https://godoc.org/github.com/posener/cmd?status.svg)](http://godoc.org/github.com/posener/cmd)
[![goreadme](https://goreadme.herokuapp.com/badge/posener/cmd.svg)](https://goreadme.herokuapp.com)

cmd is a minimalistic library that enables easy sub commands with the standard `flag` library.

This library extends the standard library `flag` package to support sub commands and more
features in a minimalistic and idiomatic API.

Features:

[x] Sub commands.

[x] Automatic bash completion.

[x] Flag values definition and check.

[x] Explicit positional arguments definition.

[x] Automatic usage text.

#### Usage

Define a root command object using the `New` function.
This object exposes the standard library's `flag.FlagSet` API, which enables adding flags in the
standard way.
Additionally, this object exposes the `SubCommand` method, which returns another command object.
This objects also exposing the same API, enabling definition of flags and nested sub commands.
The root object then have to be called with the `Parse` method, similarly to
the `flag.Parse` call.

#### Principles

* Minimalistic and `flag`-like.

* Any flag that is defined in the base command will be reflected in all of its sub commands.

* When user types the command, it starts from the command and sub commands, only then types the
flags and then the positional arguments:

```go
[command] [sub commands...] [flags...] [positional args...]
```

* When a command defines positional arguments, all its sub commands has these positional
arguments and thus can't define their own positional arguments.

* When flag configuration is wrong, the program will panic.

#### Examples

Definition and usage of sub commands and sub commands flags.

```golang
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

// Definition and usage of sub commands and sub commands flags.
func main() {
	// Parse command line arguments.
	root.ParseArgs("cmd", "sub1", "-flag1", "value")

	// Check which sub command was choses by the user.
	switch {
	case sub1.Parsed():
		fmt.Printf("Called sub1 with flag: %s", *flag1)
	case sub2.Parsed():
		fmt.Printf("Called sub2 with flag: %d", *flag2)
	}
}

```

##### Values

An example that shows how to use advanced configuration of flags and positional arguments using
the predict package.

```golang
package main

import (
	"fmt"
	"github.com/posener/cmd"
	"github.com/posener/complete/v2/predict"
)

func main() {
	var (
		root = cmd.New()
		// Define a flag with valid values 'foo' and 'bar', and enforce the values by `OptCheck()`.
		// The defined values will be used for bash completion, and since the OptCheck was set, the
		// flag value will be checked during the parse call.
		flag1 = root.String("flag1", "", "first flag", predict.OptValues("foo", "bar"), predict.OptCheck())
		// Define a flag to accept a valid Go file path. Choose to enforce the valid path using the
		// `OptCheck` function. The file name will also be completed in the bash completion
		// processes.
		file = root.String("file", "", "file path", predict.OptPredictor(predict.Files("*.go")), predict.OptCheck())
		// Positional arguments should be explicitly defined. Define positional arguments with valid
		// values of 'baz' and 'buzz', and choose not to enforce these values by not calling
		// `OptCheck`. These values will also be completed in the bash completion process.
		args = root.Args("[args...]", "positional arguments", predict.OptValues("baz", "buzz"))
	)

	// Parse fake command line arguments.
	root.ParseArgs("cmd", "-flag1", "foo", "-file", "cmd.go", "buz", "bazz")

	// Test:

	fmt.Println(*flag1, *file, *args)
}

```

 Output:

```
foo cmd.go [buz bazz]

```

##### Args

In the cmd package, positional arguments should be explicitly defined. They are defined using the
`Args` or `ArgsVar` methods.

```golang
package main

import (
	"fmt"
	"github.com/posener/cmd"
)

func main() {
	var (
		root = cmd.New()
		// Positional arguments should be defined as any other flag.
		args = root.Args("[args...]", "positional arguments for command line")
	)

	// Parse fake command line arguments.
	root.ParseArgs("cmd", "v1", "v2", "v3")

	// Test:

	fmt.Println(*args)
}

```

 Output:

```
[v1 v2 v3]

```

##### ArgsFn

An example of how to parse positional arguments using a custom function. It enables the advantage
of using named variables such as `src` and `dst` as opposed to args[0] and args[1].

```golang
package main

import (
	"fmt"
	"github.com/posener/cmd"
)

func main() {
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
	root.ParseArgs("cmd", "from.txt", "to.txt")

	// Test:

	fmt.Println(src, dst)
}

```

 Output:

```
from.txt to.txt

```

##### ArgsInt

An example of defining int positional arguments.

```golang
package main

import (
	"fmt"
	"github.com/posener/cmd"
)

func main() {
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
	root.ParseArgs("cmd", "10", "20", "30")

	// Test:

	sum := 0
	for _, n := range args {
		sum += n
	}
	fmt.Println(sum)
}

```

 Output:

```
60

```

##### ArgsN

An example of defining an exact number of positional arguments.

```golang
package main

import (
	"fmt"
	"github.com/posener/cmd"
)

func main() {
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
	root.ParseArgs("cmd", "from.txt", "to.txt")

	// Test:

	fmt.Println(args)
}

```

 Output:

```
[from.txt to.txt]

```


---

Created by [goreadme](https://github.com/apps/goreadme)
