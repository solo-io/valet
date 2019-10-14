package cmd

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"go.uber.org/zap"
	"os/exec"
	"strings"
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
}

func (c *Command) Run() error {
	out, err := c.Output()
	if err != nil {
		contextutils.LoggerFrom(context.TODO()).Errorw("Error running command", zap.Error(err), zap.String("out", out))
	}
	return err
}

func (c *Command) Output() (string, error) {
	cmd := exec.Command(c.Name, c.Args...)
	cmd.Stdin = strings.NewReader(c.StdIn)
	bytes, err := cmd.Output()
	if err != nil {
		return "", CommandError(err)
	}
	return string(bytes), nil
}

func (c *Command) WithStdIn(stdIn string) *Command {
	c.StdIn = stdIn
	return c
}

func (c *Command) With(args ...string) *Command {
	c.Args = append(c.Args, args...)
	return c
}

func NewCommand(cmd string, args ...string) *Command {
	return &Command{
		Name:  cmd,
		Args: args,
	}
}

