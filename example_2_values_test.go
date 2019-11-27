package cmd_test

import (
	"fmt"

	"github.com/posener/cmd"
	"github.com/posener/complete/v2/predict"
)

// An example that shows how to use advanced configuration of flags and positional arguments using
// the predict package.
func Example_values() {
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
	root.Parse([]string{"cmd", "-flag1", "foo", "-file", "cmd.go", "buz", "bazz"})

	// Test:

	fmt.Println(*flag1, *file, *args)
	// Output: foo cmd.go [buz bazz]
}
