# subcmd

[![Build Status](https://travis-ci.org/posener/subcmd.svg?branch=master)](https://travis-ci.org/posener/subcmd)
[![codecov](https://codecov.io/gh/posener/subcmd/branch/master/graph/badge.svg)](https://codecov.io/gh/posener/subcmd)
[![GoDoc](https://godoc.org/github.com/posener/subcmd?status.svg)](http://godoc.org/github.com/posener/subcmd)
[![goreadme](https://goreadme.herokuapp.com/badge/posener/subcmd.svg)](https://goreadme.herokuapp.com)

subcmd is a minimalistic library that enables easy sub commands with the standard `flag` library.

Define a root command object using the `New` function.
This object exposes the standard library's `flag.FlagSet` API, which enables adding flags in the
standard way.
Additionally, this object exposes the `SubCommand` method, which returns another command object.
This objects also exposing the same API, enabling definition of flags and nested sub commands.

The root object then have to be called with the `Parse` or `ParseArgs` methods, similarly to
the `flag.Parse` call.

The usage is automatically configured to show both sub commands and flags.

Automatic bash completion is enabled.

#### Principles

* Minimalistic and `flag`-like.

* Any flag that is defined in the base command will be reflected in all of its sub commands.

* When user types the command, it starts from the command and sub commands, only then types the
flags and then the positional arguments:

```go
[command] [sub commands...] [flags...] [positional args...]
```

* Positional arguments are as any other flag: their number and type should be enforced and
checked.

* When a command that defined positional arguments, all its sub commands has these positional
arguments and thus can't define their own positional arguments.

* Usage format is standard, programs can't define their own format.

* When flag configuration is wrong, the program will panic when starts. Most of them in flag
definition stage, and not after flag parsing stage.

#### Examples

Definition and usage of sub commands and sub commands flags.

```golang
package main

import (
	"fmt"

	"github.com/posener/subcmd"
)

var (
	// Define a cmd command. Some options can be set using the `Opt*` functions. It returns a
	// `*Cmd` object.
	cmd = subcmd.New()
	// The `*Cmd` object can be used as the standard library `flag.FlagSet`.
	flag0 = cmd.String("flag0", "", "root string flag")

	// From each command object, a sub command can be created. This can be done recursively.
	sub1 = cmd.SubCommand("sub1", "first sub command")
	// Each sub command can have flags attached.
	flag1 = sub1.String("flag1", "", "sub1 string flag")

	sub2  = cmd.SubCommand("sub2", "second sub command")
	flag2 = sub1.Int("flag2", 0, "sub2 int flag")
)

// Definition and usage of sub commands and sub commands flags.
func main() {
	// In the example we use `Parse()` for a given list of command line arguments. This is useful
	// for testing, but should be replaced with `cmd.ParseArgs()` in `main()`
	cmd.Parse([]string{"cmd", "sub1", "-flag1", "value"})

	// Usually the program should switch over the sub commands. The chosen sub command will return
	// true for the `Parsed()` method.
	switch {
	case sub1.Parsed():
		fmt.Printf("Called sub1 with flag: %s", *flag1)
	case sub2.Parsed():
		fmt.Printf("Called sub2 with flag: %d", *flag2)
	}
}

```

##### Values

Flags and positional arguments can be defined with valid values. It is also possible to enable
the check for valid on parsing times. When setting valid values, they will be completed by the
bash completion system.

```golang
package main

import (
	"fmt"
	"github.com/posener/complete/v2/predict"
	"github.com/posener/subcmd"
)

func main() {
	// Should be defined in global `var`.
	var (
		cmd = subcmd.New()
		// Define a flag with valid values 'foo' and 'bar', and enforce the values by `OptCheck()`.
		flag1 = cmd.String("flag1", "", "first flag", predict.OptValues("foo", "bar"), predict.OptCheck())
		// Define a flag with valid values of Go file names.
		file = cmd.String("file", "", "file path", predict.OptPredictor(predict.Files("*")), predict.OptCheck())
		// Define positional arguments with valid values 'baz' and 'buzz', and choose not to enforce
		// the check by not calling `OptCheck`.
		args = cmd.Args("[args...]", "positional arguments", predict.OptValues("baz", "buzz"))
	)

	// Should be in `main()`.
	cmd.Parse([]string{"cmd", "-flag1", "foo", "-file", "subcmd.go", "buz", "bazz"})

	// Test:

	fmt.Println(*flag1, *file, *args)
}

```

 Output:

```
foo subcmd.go [buz bazz]

```

##### Args

Usage of positional arguments. If a program accepts positional arguments it must declare it using
the `Args()` or the `ArgsVar()` methods. Positional arguments can be also defined on sub
commands.

```golang
package main

import (
	"fmt"
	"github.com/posener/subcmd"
)

func main() {
	// Should be defined in global `var`.
	var (
		cmd = subcmd.New()
		// Positional arguments can be defined as any other flag.
		args = cmd.Args("[args...]", "positional arguments for command line")
	)

	// Should be in `main()`.
	cmd.Parse([]string{"cmd", "v1", "v2", "v3"})

	// Test:

	fmt.Println(*args)
}

```

 Output:

```
[v1 v2 v3]

```

##### ArgsFn

Usage of positional arguments with a conversion function.

```golang
package main

import (
	"fmt"
	"github.com/posener/subcmd"
)

func main() {
	// Should be defined in global `var`.
	var (
		cmd      = subcmd.New()
		src, dst string
	)

	// A function that convert the positional arguments to the program variables.
	argsFn := func(args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("expected src and dst, got %d arguments", len(args))
		}
		src, dst = args[0], args[1]
		return nil
	}

	// Should be in `init()`.
	cmd.ArgsVar(subcmd.ArgsFn(argsFn), "[src] [dst]", "positional arguments for command line")

	// Should be in `main()`.
	cmd.Parse([]string{"cmd", "from.txt", "to.txt"})

	// Test:

	fmt.Println(src, dst)
}

```

 Output:

```
from.txt to.txt

```

##### ArgsInt

Usage of positional arguments of a specific type.

```golang
package main

import (
	"fmt"
	"github.com/posener/subcmd"
)

func main() {
	// Should be defined in global `var`.
	var (
		cmd = subcmd.New()
		// Define positional arguments of type integer.
		args subcmd.ArgsInt
	)

	// Should be in `init()`.
	cmd.ArgsVar(&args, "[int...]", "numbers to sum")

	// Should be in `main()`.
	cmd.Parse([]string{"cmd", "10", "20", "30"})

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

Usage of positional arguments with exact number of arguments.

```golang
package main

import (
	"fmt"
	"github.com/posener/subcmd"
)

func main() {
	// Should be defined in global `var`.
	var (
		cmd = subcmd.New()
		// Define arguments with cap=2 will ensure that the number of arguments is always 2.
		args = make(subcmd.ArgsStr, 2)
	)

	// Should be in `init()`.
	cmd.ArgsVar(&args, "[src] [dst]", "positional arguments for command line")

	// Should be in `main()`.
	cmd.Parse([]string{"cmd", "from.txt", "to.txt"})

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
