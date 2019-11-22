package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/go-utils/errors"
)

type Kind struct {
	cmd *Command
}

func (k *Kind) With(args ...string) *Kind {
	k.cmd = k.cmd.With(args...)
	return k
}

func (k *Kind) SwallowError() *Kind {
	k.cmd.SwallowErrorLog = true
	return k
}

func (k *Kind) Cmd() *Command {
	return k.cmd
}

func (k *Kind) Name(name string) *Kind {
	return k.With(fmt.Sprintf("--name=%s", name))
}

func (k *Kind) IsRunning(ctx context.Context, runner Runner, name string) (bool, error) {
	output, err := runner.Output(ctx, k.With("get", "clusters").SwallowError().Cmd())
	if err != nil {
		return false, errors.Wrapf(err, output)

	}
	if strings.Contains(output, name) {
		return true, nil
	}
	return false, nil
}

func (k *Kind) CreateCluster(ctx context.Context, runner Runner, name string) error {
	return runner.RunInShell(ctx, k.With("create", "cluster").Name(name).Cmd())
}

func (k *Kind) DeleteCluster(ctx context.Context, runner Runner, name string) error {
	return runner.RunInShell(ctx, k.With("delete", "cluster").Name(name).Cmd())
}
