package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/go-utils/errors"
)

type EksCtl struct {
	cmd *Command
}

func (e *EksCtl) Cmd() *Command {
	return e.cmd
}

func (e *EksCtl) With(args ...string) *EksCtl {
	e.cmd = e.cmd.With(args...)
	return e
}

func (e *EksCtl) SwallowError() *EksCtl {
	e.cmd.SwallowErrorLog = true
	return e
}

func (e *EksCtl) Region(region string) *EksCtl {
	if region != "" {
		return e.With(fmt.Sprintf("--region=%s", region))
	}
	return e
}

func (e *EksCtl) WithName(name string) *EksCtl {
	return e.With(name)
}

func (e *EksCtl) Name(name string) *EksCtl {
	if name != "" {
		return e.With(fmt.Sprintf("--name=%s", name))
	}
	return e
}

func (e *EksCtl) GetCredentials() *EksCtl {
	return e.With("utils", "write-kubeconfig")
}

func (e *EksCtl) GetCluster() *EksCtl {
	return e.With("get", "cluster")
}

func (e *EksCtl) DeleteCluster(ctx context.Context, name, region string, runner Runner) error {
	streamHandler, err := runner.Stream(ctx, e.With("delete", "cluster").Region(region).WithName(name).Cmd())
	if err != nil {
		return err
	}
	inputErr := errors.New("unable to delete cluster resources")
	return streamHandler.StreamHelper(inputErr)
}

func (e *EksCtl) CreateCluster(ctx context.Context, name, region string, runner Runner) error {
	streamHandler, err := runner.Stream(ctx, e.With("create", "cluster").Region(region).WithName(name).Cmd())
	if err != nil {
		return err
	}
	inputErr := errors.New("unable to create cluster resources")
	return streamHandler.StreamHelper(inputErr)
}

func (e *EksCtl) IsRunning(ctx context.Context, name, region string, runner Runner) (bool, error) {
	output, err := runner.Output(ctx, e.GetCluster().Region(region).Name(name).SwallowError().Cmd())
	if err != nil {
		if strings.Contains(output, "ResourceNotFoundException: No cluster found for name") {
			return false, nil
		}
		return false, errors.Wrapf(err, output)
	}
	return true, nil
}

func (e *EksCtl) WriteKubeConfig(ctx context.Context, name, region string, runner Runner) error {
	output, err := runner.Output(ctx, e.GetCredentials().Region(region).Name(name).With("--auto-kubeconfig").SwallowError().Cmd())
	if err != nil {
		return errors.Wrapf(err, output)
	}
	return nil
}
