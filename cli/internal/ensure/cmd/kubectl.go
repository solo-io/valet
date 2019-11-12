package cmd

import (
	"context"
	"fmt"
)

type Kubectl struct {
	cmd *Command
}

func (k *Kubectl) SwallowErrorLog(swallow bool) *Kubectl {
	k.cmd.SwallowErrorLog = swallow
	return k
}

func (k *Kubectl) Cmd() *Command {
	return k.cmd
}

func (k *Kubectl) With(args ...string) *Kubectl {
	k.cmd = k.cmd.With(args...)
	return k
}

func (k *Kubectl) WithStdIn(stdIn string) *Kubectl {
	k.cmd.StdIn = stdIn
	return k
}

func (k *Kubectl) File(file string) *Kubectl {
	return k.With("-f", file)
}

func (k *Kubectl) WithName(name string) *Kubectl {
	return k.With(name)
}

func (k *Kubectl) Namespace(ns string) *Kubectl {
	return k.With("-n", ns)
}

func (k *Kubectl) Context(context string) *Kubectl {
	return k.With("--context", context)
}

func (k *Kubectl) DryRun() *Kubectl {
	return k.With("--dry-run")
}

func (k *Kubectl) OutYaml() *Kubectl {
	return k.With("-oyaml")
}

func (k *Kubectl) IgnoreNotFound() *Kubectl {
	return k.With("--ignore-not-found")
}

func (k *Kubectl) Create(typeToCreate string) *Kubectl {
	return k.With("create", typeToCreate)
}

func (k *Kubectl) Delete(typeToDelete string) *Kubectl {
	return k.With("delete", typeToDelete)
}

func (k *Kubectl) Apply() *Kubectl {
	return k.With("apply")
}

func (k *Kubectl) ApplyStdIn(stdIn string) *Kubectl {
	return k.Apply().File("-").WithStdIn(stdIn)
}

func (k *Kubectl) ApplyFile(path string) *Kubectl {
	return k.Apply().File(path)
}

func (k *Kubectl) Redact(unredacted, redacted string) *Kubectl {
	if k.cmd.Redactions == nil {
		k.cmd.Redactions = make(map[string]string)
	}
	k.cmd.Redactions[unredacted] = redacted
	return k
}

func (k *Kubectl) UseContext(context string) *Kubectl {
	return k.With("config", "use-context", context)
}

func (k *Kubectl) CurrentContext() *Kubectl {
	return k.With("config", "current-context")
}

func (k *Kubectl) DryRunAndApply(ctx context.Context, runner Runner) error {
	out, err := runner.Output(ctx, k.DryRun().OutYaml().Cmd())
	if err != nil {
		return err
	}
	return runner.Run(ctx, New().Kubectl().ApplyStdIn(out).Cmd())
}

func (k *Kubectl) JsonPatch(jsonPatch string) *Kubectl {
	return k.With("--type=json", jsonPatch)
}

func (k *Kubectl) DeleteFile(path string) *Kubectl {
	return k.With("delete", "-f", path)
}

func (k *Kubectl) DeleteStdIn(stdIn string) *Kubectl {
	return k.DeleteFile("-").WithStdIn(stdIn)
}

func (k *Kubectl) OutJsonpath(jsonpath string) *Kubectl {
	return k.With(fmt.Sprintf("-o=jsonpath=%s", jsonpath))
}

func (k *Kubectl) GetServiceIP(ctx context.Context, namespace, name string, runner Runner) (string, error) {
	cmd := k.With("get", "svc", name).Namespace(namespace)
	cmd = cmd.OutJsonpath("{ .status.loadBalancer.ingress[0].ip }")
	return runner.Output(ctx, cmd.Cmd())
}
