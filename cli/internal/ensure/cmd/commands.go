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

	Redactions      map[string]string
	SwallowErrorLog bool
}

func (c *Command) Run(ctx context.Context) error {
	_, err := c.Output(ctx)
	return err
}

func (c *Command) Output(ctx context.Context) (string, error) {
	c.logCommand(ctx)
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdin = strings.NewReader(c.StdIn)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		if !c.SwallowErrorLog {
			contextutils.LoggerFrom(ctx).Errorw("error running command", zap.Error(err), zap.String("out", string(bytes)))
		}
		return "", CommandError(err)
	}
	return string(bytes), nil
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
	contextutils.LoggerFrom(ctx).Infow("running command",
		zap.String("command", command))
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
