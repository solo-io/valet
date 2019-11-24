package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/solo-io/go-utils/errors"
	"k8s.io/client-go/tools/clientcmd"
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

func (k *Kind) KubeConfig(kubeConfig string) *Kind {
	k.cmd.Env[clientcmd.RecommendedConfigPathEnvVar] = kubeConfig
	return k
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
	streamHandler, err := runner.Stream(ctx, k.With("create", "cluster").Name(name).Cmd())
	if err != nil {
		return err
	}
	done := make(chan struct{})
	go func() {
		ticker := time.Tick(time.Second * 5)
		for {
			select {
			case <-done:
				return
			case <-ticker:
				Stdout(ctx).Println("creating kind cluster")
			}
		}
	}()
	stderr, _ := ioutil.ReadAll(streamHandler.Stderr)
	err = streamHandler.WaitFunc()
	done <- struct{}{}
	if err != nil {
		Stderr(ctx).Println(fmt.Sprintf("%s\n", stderr))
		return err
	}
	Stdout(ctx).Println("successfully created kind cluster")
	return nil
}

func (k *Kind) DeleteCluster(ctx context.Context, runner Runner, name string) error {
	streamHandler, err := runner.Stream(ctx, k.With("delete", "cluster").Name(name).Cmd())
	if err != nil {
		return err
	}
	inputErr := errors.New("unable to delete cluster resources")
	return streamHandler.StreamHelper(ctx, inputErr)
}
