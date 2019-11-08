package subcmd

import (
	"flag"
	"time"
)

func (c *SubCmd) checkNewFlag() {
	// If subcommands were set, new flag can't be set anymore.
	if len(c.sub) > 0 {
		panic("flags must be defined before defining sub commands")
	}
}

func (c *SubCmd) checkNewArgs() {
	// If subcommands were set, positional arguments can't be set anymore.
	if len(c.sub) > 0 {
		panic("positional args must be defined before defining sub commands")
	}
}

func (c *SubCmd) Parsed() bool { return c.flagSet.Parsed() }

func (c *SubCmd) Set(name, value string) error {
	return c.flagSet.Set(name, value)
}

func (c *SubCmd) Var(value flag.Value, name string, usage string) {
	c.checkNewFlag()
	c.flagSet.Var(value, name, usage)
}

func (c *SubCmd) Visit(fn func(*flag.Flag)) {
	c.flagSet.Visit(fn)
}

func (c *SubCmd) VisitAll(fn func(*flag.Flag)) {
	c.flagSet.VisitAll(fn)
}

func (c *SubCmd) String(name string, value string, usage string) *string {
	c.checkNewFlag()
	return c.flagSet.String(name, value, usage)
}

func (c *SubCmd) StringVar(p *string, name string, value string, usage string) {
	c.checkNewFlag()
	c.flagSet.StringVar(p, name, value, usage)
}

func (c *SubCmd) Bool(name string, value bool, usage string) *bool {
	c.checkNewFlag()
	return c.flagSet.Bool(name, value, usage)
}

func (c *SubCmd) BoolVar(p *bool, name string, value bool, usage string) {
	c.checkNewFlag()
	c.flagSet.BoolVar(p, name, value, usage)
}

func (c *SubCmd) Int(name string, value int, usage string) *int {
	c.checkNewFlag()
	return c.flagSet.Int(name, value, usage)
}

func (c *SubCmd) IntVar(p *int, name string, value int, usage string) {
	c.checkNewFlag()
	c.flagSet.IntVar(p, name, value, usage)
}

func (c *SubCmd) Int64(name string, value int64, usage string) *int64 {
	c.checkNewFlag()
	return c.flagSet.Int64(name, value, usage)
}

func (c *SubCmd) Int64Var(p *int64, name string, value int64, usage string) {
	c.checkNewFlag()
	c.flagSet.Int64Var(p, name, value, usage)
}

func (c *SubCmd) Float64(name string, value float64, usage string) *float64 {
	c.checkNewFlag()
	return c.flagSet.Float64(name, value, usage)
}

func (c *SubCmd) Float64Var(p *float64, name string, value float64, usage string) {
	c.checkNewFlag()
	c.flagSet.Float64Var(p, name, value, usage)
}

func (c *SubCmd) Uint(name string, value uint, usage string) *uint {
	c.checkNewFlag()
	return c.flagSet.Uint(name, value, usage)
}

func (c *SubCmd) UintVar(p *uint, name string, value uint, usage string) {
	c.checkNewFlag()
	c.flagSet.UintVar(p, name, value, usage)
}

func (c *SubCmd) Uint64(name string, value uint64, usage string) *uint64 {
	c.checkNewFlag()
	return c.flagSet.Uint64(name, value, usage)
}

func (c *SubCmd) UintVar64(p *uint64, name string, value uint64, usage string) {
	c.checkNewFlag()
	c.flagSet.Uint64Var(p, name, value, usage)
}

func (c *SubCmd) Duration(name string, value time.Duration, usage string) *time.Duration {
	c.checkNewFlag()
	return c.flagSet.Duration(name, value, usage)
}

func (c *SubCmd) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	c.checkNewFlag()
	c.flagSet.DurationVar(p, name, value, usage)
}
