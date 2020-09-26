package cmd

import (
	"testing"

	"github.com/posener/complete/v2"
)

func TestComplete(t *testing.T) {
	t.Parallel()

	comp := (*completer)(newTestCmd().SubCmd)

	tests := []struct {
		line        string
		completions []string
	}{
		// Check completion of sub commands.
		{line: "su", completions: []string{"sub1", "sub2"}},
		// Check completion of flag names.
		{line: "sub1 sub1 -f", completions: []string{"-flag1", "-flag0", "-flag11"}},
		// Check completion of flag values.
		{line: "sub1 sub1 -flag1 ", completions: []string{"foo", "bar"}},
		// Check completion for positional arguments.
		{line: "sub2 ", completions: []string{"-flag0", "-h", "one", "two"}},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			complete.Test(t, comp, tt.line, tt.completions)
		})
	}
}
