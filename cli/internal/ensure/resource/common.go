package resource

import (
	"context"
	"reflect"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

type Resource interface {
	Ensure(ctx context.Context, inputs render.InputParams, command cmd.Factory) error
	Teardown(ctx context.Context, inputs render.InputParams, command cmd.Factory) error
}

func EnsureAll(ctx context.Context, input render.InputParams, command cmd.Factory, resources ...Resource) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		if input.Step {
			if err := cmd.PromptPressAnyKeyToContinue(); err != nil {
				return err
			}
		}
		if err := resource.Ensure(ctx, input, command); err != nil {
			return err
		}
	}
	return nil
}

func EnsureFirst(ctx context.Context, input render.InputParams, command cmd.Factory, resources ...Resource) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		return resource.Ensure(ctx, input, command)
	}
	return nil
}

func TeardownAll(ctx context.Context, input render.InputParams, command cmd.Factory, resources ...Resource) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		if err := resource.Teardown(ctx, input, command); err != nil {
			return err
		}
	}
	return nil
}

func TeardownFirst(ctx context.Context, input render.InputParams, command cmd.Factory, resources ...Resource) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		return resource.Teardown(ctx, input, command)
	}
	return nil
}
