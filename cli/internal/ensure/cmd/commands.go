package cmd

import (
	"context"
	"os/exec"
	"strings"

	"github.com/solo-io/go-utils/errors"
)

const (
	Redacted = "REDACTED"
	Empty    = "EMPTY"
)

var (
	CommandError = func(err error) error {
		return errors.Wrapf(err, "command error")
	}
)

type Command struct {
	Name  string
	Args  []string
	StdIn string

	PrintCommands   bool
	Redactions      map[string]string
	SwallowErrorLog bool
	CommandRunner   CommandRunner
}

func (c *Command) Run(ctx context.Context) error {
	runner := c.CommandRunner
	if runner == nil {
		runner = &commandRunner{}
	}
	return runner.Run(ctx, c)
}

func (c *Command) Output(ctx context.Context) (string, error) {
	runner := c.CommandRunner
	if runner == nil {
		runner = &commandRunner{}
	}
	return runner.Output(ctx, c)
}

type CommandRunner interface {
	Run(ctx context.Context, c *Command) error
	Output(ctx context.Context, c *Command) (string, error)
}

type commandRunner struct {
}

func (r *commandRunner) Run(ctx context.Context, c *Command) error {
	_, err := r.Output(ctx, c)
	return err
}

func (r *commandRunner) Output(ctx context.Context, c *Command) (string, error) {
	c.logCommand(ctx)
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdin = strings.NewReader(c.StdIn)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		if !c.SwallowErrorLog {
			Stderr().Println("Error running command: %s", err.Error())
			Stderr().Println(string(bytes))
		}
		err = CommandError(err)
	}
	return string(bytes), err
}

func (c *Command) logCommand(ctx context.Context) {
	var parts []string
	parts = append(parts, c.Name)
	for _, arg := range c.Args {
		processed := arg
		for unredacted, redacted := range c.Redactions {
			if arg == unredacted {
				processed = redacted
				break
			}
		}
		if arg == "" {
			processed = Empty
		} else if processed == "" {
			processed = Redacted
		}
		parts = append(parts, processed)
	}
	command := strings.Join(parts, " ")
	if c.PrintCommands {
		Stdout().Println("Running command: %s", command)
	}
}

func (c *Command) WithStdIn(stdIn string) *Command {
	c.StdIn = stdIn
	return c
}

func (c *Command) With(args ...string) *Command {
	c.Args = append(c.Args, args...)
	return c
}

func (c *Command) Redact(unredacted, redacted string) *Command {
	if c.Redactions == nil {
		c.Redactions = make(map[string]string)
	}
	c.Redactions[unredacted] = redacted
	return c
}
