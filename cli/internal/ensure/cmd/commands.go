package cmd

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"go.uber.org/zap"
	"os/exec"
	"strings"
)

const (
	Redacted = "REDACTED"
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

	RedactedArgs []string
}

func (c *Command) Run(ctx context.Context) error {
	_, err := c.Output(ctx)
	return err
}

func (c *Command) Output(ctx context.Context) (string, error) {
	contextutils.LoggerFrom(ctx).Infow("Running command")
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdin = strings.NewReader(c.StdIn)
	bytes, err := cmd.Output()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error running command", zap.Error(err), zap.String("out", string(bytes)))
		return "", CommandError(err)
	}
	return string(bytes), nil
}

func (c *Command) logCommand(ctx context.Context) {
	var parts []string
	parts = append(parts, c.Name)
	for _, arg := range c.Args {
		logArg := true
		for _, redacted := range c.RedactedArgs {
			if arg == redacted {
				logArg = false
				break
			}
		}
		if logArg {
			parts = append(parts, arg)
		} else {
			parts = append(parts, Redacted)
		}
	}
	command := strings.Join(parts, " ")
	contextutils.LoggerFrom(ctx).Infow("running command",
		zap.String("command", command),
		zap.String("stdIn", c.StdIn))
}

func (c *Command) WithStdIn(stdIn string) *Command {
	c.StdIn = stdIn
	return c
}

func (c *Command) With(args ...string) *Command {
	c.Args = append(c.Args, args...)
	return c
}

func (c *Command) Redact(args ...string) *Command {
	c.RedactedArgs = append(c.RedactedArgs, args...)
	return c
}
