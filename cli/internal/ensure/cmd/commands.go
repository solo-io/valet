package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/solo-io/go-utils/errors"
)

//go:generate mockgen -destination ./mocks/command_runner_mock.go github.com/solo-io/valet/cli/internal/ensure/cmd Runner

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
}

type Runner interface {
	Run(ctx context.Context, c *Command) error
	Output(ctx context.Context, c *Command) (string, error)
	Stream(ctx context.Context, c *Command) (*CommandStreamHandler, error)
	Request(ctx context.Context, req *http.Request) (string, int, error)
}

type commandRunner struct{}

func DefaultCommandRunner() Runner {
	return &commandRunner{}
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
			Stderr().Println("STDIN: %s", c.StdIn)
			Stderr().Println(string(bytes))
		}
		err = CommandError(err)
	}
	return string(bytes), err
}

type CommandStreamHandler struct {
	WaitFunc func() error
	Stdout   io.Reader
	Stderr   io.Reader
}

func (c *CommandStreamHandler) StreamHelper(inputErr error) error {
	go func() {
		stdoutScanner := bufio.NewScanner(c.Stdout)
		for stdoutScanner.Scan() {
			Stdout().Println(stdoutScanner.Text())
		}
		if err := stdoutScanner.Err(); err != nil {
			Stderr().Println("reading stdout from current command context:", err)
		}
	}()
	stderr, _ := ioutil.ReadAll(c.Stderr)
	if err := c.WaitFunc(); err != nil {
		Stderr().Println(fmt.Sprintf("%s\n", stderr))
		return inputErr
	}
	return nil
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

func (c *commandRunner) Request(ctx context.Context, req *http.Request) (string, int, error) {
	httpClient := &http.Client{
		Timeout: time.Second * 1,
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	p := new(bytes.Buffer)
	_, err = io.Copy(p, resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", 0, err
	}
	return p.String(), resp.StatusCode, nil
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
