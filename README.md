# subcmd

[![Build Status](https://travis-ci.org/posener/subcmd.svg?branch=master)](https://travis-ci.org/posener/subcmd)
[![codecov](https://codecov.io/gh/posener/subcmd/branch/master/graph/badge.svg)](https://codecov.io/gh/posener/subcmd)
[![GoDoc](https://godoc.org/github.com/posener/subcmd?status.svg)](http://godoc.org/github.com/posener/subcmd)
[![goreadme](https://goreadme.herokuapp.com/badge/posener/subcmd.svg)](https://goreadme.herokuapp.com)

subcmd is a minimalistic library that enables easy sub commands with the standard `flag` library.

Define a `root` command object using the `Root` function.
This object exposes the standard library's `flag.FlagSet` API, which enables adding flags in the
standard way.
Additionally, this object exposes the `SubCommand` method, which returns another command object.
This objects also exposing the same API, enabling definition of flags and nested sub commands.

The root object then have to be called with the `Parse` or `ParseArgs` methods, similarly to
the `flag.Parse` call.

The usage is automatically configured to show both sub commands and flags.

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
	// Define a root command. Some options can be set using the `Opt*` functions. It returns a
	// `*Cmd` object.
	root = subcmd.Root()
	// The `*Cmd` object can be used as the standard library `flag.FlagSet`.
	flag0 = root.String("flag0", "", "root stringflag")
	// From each command object, a sub command can be created. This can be done recursively.
	sub1 = root.SubCommand("sub1", "first sub command")
	// Each sub command can have flags attached.
	flag1 = sub1.String("flag1", "", "sub1 string flag")
	sub2  = root.SubCommand("sub2", "second sub command")
	flag2 = sub1.Int("flag2", 0, "sub2 int flag")
)

// Definition and usage of sub commands and sub commands flags.
func main() {
	// In the example we use `Parse()` for a given list of command line arguments. This is useful
	// for testing, but should be replaced with `root.ParseArgs()` in `main()`
	root.Parse([]string{"cmd", "sub1", "-flag1", "value"})

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
		root = subcmd.Root()
		// Positional arguments can be defined as any other flag.
		args = root.Args("[args...]", "positional arguments for command line")
	)

	// Should be in `main()`.
	root.Parse([]string{"cmd", "v1", "v2", "v3"})

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
		root     = subcmd.Root()
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
	root.ArgsVar(subcmd.ArgsFn(argsFn), "[src] [dst]", "positional arguments for command line")

	// Should be in `main()`.
	root.Parse([]string{"cmd", "from.txt", "to.txt"})

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
		root = subcmd.Root()
		// Define positional arguments of type integer.
		args subcmd.ArgsInt
	)

	// Should be in `init()`.
	root.ArgsVar(&args, "[int...]", "numbers to sum")

	// Should be in `main()`.
	root.Parse([]string{"cmd", "10", "20", "30"})

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
		root = subcmd.Root()
		// Define arguments with cap=2 will ensure that the number of arguments is always 2.
		args = make(subcmd.ArgsStr, 2)
	)

	// Should be in `init()`.
	root.ArgsVar(&args, "[src] [dst]", "positional arguments for command line")

	// Should be in `main()`.
	root.Parse([]string{"cmd", "from.txt", "to.txt"})

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
