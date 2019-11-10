package subcmd

import (
	"flag"
	"github.com/posener/complete/v2"
)

type completer SubCmd

func (c *completer) SubCmdList() []string {
	subcmd := (*SubCmd)(c)
	return subcmd.subNames()
}

func (c *completer) SubCmdGet(cmd string) complete.Completer {
	if c.sub[cmd] == nil {
		return nil
	}
	return (*completer)(c.sub[cmd])
}

func (c *completer) FlagList() []string {
	if len(c.sub) != 0 {
		return nil
	}
	var flags []string
	c.flagSet.VisitAll(func(f *flag.Flag) {
		flags = append(flags, f.Name)
	})
	return flags
}

func (c *completer) FlagGet(flag string) complete.Predictor {
	f := c.flagSet.Lookup(flag)
	if f == nil {
		return nil
	}
	if p, ok := f.Value.(complete.Predictor); ok {
		return p
	}
	return nil
}

func (c *completer) ArgsGet() complete.Predictor {
	return nil
}
