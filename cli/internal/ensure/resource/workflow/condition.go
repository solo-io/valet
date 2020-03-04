package workflow

import (
	"context"
	"time"

	"github.com/solo-io/valet/cli/internal/ensure/resource/render"

	errors "github.com/rotisserie/eris"
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
	Type      string `json:"type"`
	Name      string `json:"name" valet:"template"`
	Namespace string `json:"namespace"`
	Jsonpath  string `json:"jsonpath"`
	Value     string `json:"value"`
	Timeout   string `json:"timeout" valet:"template,default=120s"`
	Interval  string `json:"interval" valet:"template,default=5s"`
}

func (c *Condition) Ensure(ctx context.Context, input render.InputParams) error {
	if err := input.RenderFields(c); err != nil {
		return err
	}
	cmd.Stdout().Println("Waiting on condition: %s.%s path %s matches %s (timeout: %s)", c.Namespace, c.Name, c.Jsonpath, c.Value, c.Timeout)
	if c.conditionMet(ctx, input) {
		return nil
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
			if c.conditionMet(ctx, input) {
				return nil
			}
		}
	}
}

func (c *Condition) conditionMet(ctx context.Context, input render.InputParams) bool {
	out, err := input.Runner().Output(ctx, c.GetConditionCmd())
	if err != nil {
		cmd.Stderr().Println("Error checking condition")
		return false
	}
	if out == c.Value {
		cmd.Stdout().Println("Condition met!")
		return true
	}
	return false
}

func (*Condition) Teardown(ctx context.Context, input render.InputParams) error {
	cmd.Stdout().Println("Skipping teardown for condition")
	return nil
}

func (c *Condition) GetConditionCmd() *cmd.Command {
	return cmd.New().Kubectl().
		With("get", c.Type, c.Name).
		Namespace(c.Namespace).
		OutJsonpath(c.Jsonpath).Cmd()
}
