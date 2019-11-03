package resource

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	DefaultTimeout = "120s"
)

var (
	ConditionNotMetError = errors.Errorf("Condition wasn't met")
)

type Condition struct {
	Type      string `yaml:"type"`
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Jsonpath  string `yaml:"jsonpath"`
	Value     string `yaml:"value"`
	Timeout   string `yaml:"timeout"`
}

func (c *Condition) Ensure(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Waiting on condition: %s.%s (%s) path %s matches %s (timeout: %s)", c.Namespace, c.Name, c.Timeout, c.Jsonpath, c.Value, c.Timeout)
	if c.Timeout == "" {
		c.Timeout = DefaultTimeout
	}
	timeoutDuration, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return err
	}
	timeout := time.After(timeoutDuration)
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-timeout:
			return ConditionNotMetError
		case <-tick:
			out, err := command.Kubectl().
				With("get", c.Type, c.Name).
				Namespace(c.Namespace).
				OutJsonpath(c.Jsonpath).Cmd().Output(ctx)
			if err != nil {
				return err
			}
			if out == c.Value {
				return nil
			}
		}
	}
}

func (*Condition) Teardown(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Skipping teardown for condition")
	return nil
}
