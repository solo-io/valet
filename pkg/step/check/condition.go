package check

import (
	"fmt"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
	"time"

	errors "github.com/rotisserie/eris"
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

func (c *Condition) GetDescription(ctx *api.WorkflowContext, values render.Values) (string, error) {
	if err := values.RenderFields(c, ctx.Runner); err != nil {
		return "", err
	}
	return fmt.Sprintf("Waiting for the following command to return '%s': %s", c.Value, c.GetCmd().ToString()), nil
}

func (c *Condition) Run(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(c, ctx.Runner); err != nil {
		return err
	}
	if c.conditionMet(ctx) {
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
			if c.conditionMet(ctx) {
				return nil
			}
		}
	}
}

func (c *Condition) GetDocs(ctx *api.WorkflowContext, values render.Values, flags render.Flags) (string, error) {
	panic("implement me")
}

func (c *Condition) conditionMet(ctx *api.WorkflowContext) bool {
	out, err := ctx.Runner.Output(c.GetCmd())
	if err != nil {
		cmd.Stderr().Println("Error checking condition: %v", err)
		return false
	}
	if out == c.Value {
		cmd.Stdout().Println("Condition met!")
		return true
	}
	return false
}

func (c *Condition) GetCmd() *cmd.Command {
	return cmd.New().Kubectl().
		With("get", c.Type, c.Name).
		Namespace(c.Namespace).
		OutJsonpath(c.Jsonpath).Cmd()
}
