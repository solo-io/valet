package resource

import (
	"context"
	"reflect"

	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
)

type Resource interface {
	Ensure(ctx context.Context, inputs render.InputParams) error
	Teardown(ctx context.Context, inputs render.InputParams) error
}

func EnsureAll(ctx context.Context, input render.InputParams, resources ...Resource) error {
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
		if err := resource.Ensure(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func EnsureFirst(ctx context.Context, input render.InputParams, resources ...Resource) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		return resource.Ensure(ctx, input)
	}
	return nil
}

func TeardownAll(ctx context.Context, input render.InputParams, resources ...Resource) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		if err := resource.Teardown(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func TeardownFirst(ctx context.Context, input render.InputParams, resources ...Resource) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		return resource.Teardown(ctx, input)
	}
	return nil
}
