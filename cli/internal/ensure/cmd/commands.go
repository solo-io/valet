package cmd

import (
	"context"
	"io"
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

func (c *Command) Stream(ctx context.Context) (*CommandStreamHandler, error) {
	runner := c.CommandRunner
	if runner == nil {
		runner = &commandRunner{}
	}
	return runner.Stream(ctx, c)
}

type CommandRunner interface {
	Run(ctx context.Context, c *Command) error
	Output(ctx context.Context, c *Command) (string, error)
	Stream(ctx context.Context, c *Command) (*CommandStreamHandler, error)
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
		return string(bytes), CommandError(err)
	}
	return string(bytes), nil
}

type CommandStreamHandler struct {
	WaitFunc func() error
	Stdout   io.Reader
	Stderr   io.Reader
}

func (r *commandRunner) Stream(ctx context.Context, c *Command) (*CommandStreamHandler, error) {
	c.logCommand(ctx)
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdin = strings.NewReader(c.StdIn)
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	return &CommandStreamHandler{
		WaitFunc: func() error {
			return cmd.Wait()
		},
		Stdout: outReader,
		Stderr: errReader,
	}, nil
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
