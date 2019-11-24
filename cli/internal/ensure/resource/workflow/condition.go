package workflow

import (
	"context"
	"time"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	DefaultConditionTimeout  = "120s"
	DefaultConditionInterval = "5s"
)

var (
	ConditionNotMetError = errors.Errorf("Condition wasn't met")
)

type Condition struct {
	Type      string `yaml:"type"`
	Name      string `yaml:"name" valet:"template"`
	Namespace string `yaml:"namespace"`
	Jsonpath  string `yaml:"jsonpath"`
	Value     string `yaml:"value"`
	Timeout   string `yaml:"timeout" valet:"template,default=120s"`
	Interval  string `yaml:"interval" valet:"template,default=5s"`
}

func (c *Condition) Ensure(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(c); err != nil {
		return err
	}
	cmd.Stdout(ctx).Println("Waiting on condition: %s.%s path %s matches %s (timeout: %s)", c.Namespace, c.Name, c.Jsonpath, c.Value, c.Timeout)
	if met, err := c.conditionMet(ctx, input); err != nil || met {
		return err
	}
	timeoutDuration, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return err
	}
	interval, err := time.ParseDuration(c.Interval)
	if err != nil {
		return err
	}
	timeout := time.After(timeoutDuration)
	tick := time.Tick(interval)
	for {
		select {
		case <-timeout:
			return ConditionNotMetError
		case <-tick:
			if met, err := c.conditionMet(ctx, input); err != nil || met {
				return err
			}
		}
	}
}

func (c *Condition) conditionMet(ctx context.Context, input render.InputParams) (bool, error) {
	out, err := input.Runner().Output(ctx, c.GetConditionCmd(input))
	if err != nil {
		cmd.Stderr(ctx).Println("Error checking condition")
		return false, err
	}
	if out == c.Value {
		cmd.Stdout(ctx).Println("Condition met!")
		return true, nil
	}
	return false, nil
}

func (*Condition) Teardown(ctx context.Context, input render.InputParams) error {
	cmd.Stdout(ctx).Println("Skipping teardown for condition")
	return nil
}

func (c *Condition) GetConditionCmd(input render.InputParams) *cmd.Command {
	return cmd.New().Kubectl().
		With("get", c.Type, c.Name).
		Namespace(c.Namespace).
		OutJsonpath(c.Jsonpath).Cmd(input.KubeConfig())
}
