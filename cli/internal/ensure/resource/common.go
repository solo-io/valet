package resource

import (
	"context"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"reflect"
)

func EnsureAll(ctx context.Context, input InputParams, command cmd.Factory, resources ...Resource) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		if err := resource.Ensure(ctx, input, command); err != nil {
			return err
		}
	}
	return nil
}

func EnsureFirst(ctx context.Context, input InputParams, command cmd.Factory, resources ...Resource) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		return resource.Ensure(ctx, input, command)
	}
	return nil
}

func TeardownAll(ctx context.Context, input InputParams, command cmd.Factory, resources ...Resource) error {
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

func TeardownFirst(ctx context.Context, input InputParams, command cmd.Factory, resources ...Resource) error {
	for _, resource := range resources {
		t := reflect.ValueOf(resource)
		if t.IsNil() {
			continue
		}
		return resource.Teardown(ctx, input, command)
	}
	return nil
}
