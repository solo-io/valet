package cmd

import "context"

type kubectl Command

func (k *kubectl) With(args ...string) *kubectl {
	k.Args = append(k.Args, args...)
	return k
}

func (k *kubectl) WithStdIn(stdIn string) *kubectl {
	k.StdIn = stdIn
	return k
}

func (k *kubectl) File(file string) *kubectl {
	return k.With("-f", file)
}

func (k *kubectl) WithName(name string) *kubectl {
	return k.With(name)
}

func (k *kubectl) Namespace(ns string) *kubectl {
	return k.With("-n", ns)
}

func (k *kubectl) DryRun() *kubectl {
	return k.With("--dry-run")
}

func (k *kubectl) OutYaml() *kubectl {
	return k.With("-oyaml")
}

func (k *kubectl) IgnoreNotFound() *kubectl {
	return k.With("--ignore-not-found")
}

func (k *kubectl) Create(typeToCreate string) *kubectl {
	return k.With("create", typeToCreate)
}

func (k *kubectl) Delete(typeToDelete string) *kubectl {
	return k.With("delete", typeToDelete)
}

func (k *kubectl) DeleteFile(path string) *kubectl {
	return k.With("delete").File(path)
}

func (k *kubectl) DeleteStdIn(stdIn string) *kubectl {
	return k.DeleteFile("-").WithStdIn(stdIn)
}

func (k *kubectl) Apply() *kubectl {
	return k.With("apply")
}

func (k *kubectl) ApplyStdIn(stdIn string) *kubectl {
	return k.Apply().File("-").WithStdIn(stdIn)
}

func (k *kubectl) ApplyFile(path string) *kubectl {
	return k.Apply().File(path)
}

func (k *kubectl) Redact(unredacted, redacted string) *kubectl {
	if k.Redactions == nil {
		k.Redactions = make(map[string]string)
	}
	k.Redactions[unredacted] = redacted
	return k
}

func (k *kubectl) Command() *Command {
	return &Command{
		Name:       k.Name,
		Args:       k.Args,
		StdIn:      k.StdIn,
		Redactions: k.Redactions,
	}
}

func (k *kubectl) Run(ctx context.Context) error {
	return k.Command().Run(ctx)
}

func (k *kubectl) UseContext(context string) *kubectl {
	return k.With("config", "use-context", context)
}

func (k *kubectl) CurrentContext() *kubectl {
	return k.With("config", "current-context")
}

func (k *kubectl) Output(ctx context.Context) (string, error) {
	return k.Command().Output(ctx)
}

func (k *kubectl) DryRunAndApply(ctx context.Context, command Factory) error {
	out, err := k.DryRun().OutYaml().Output(ctx)
	if err != nil {
		return err
	}
	return command.Kubectl().ApplyStdIn(out).Run(ctx)
}

func (k *kubectl) JsonPatch(jsonPatch string) *kubectl {
	return k.With("--type=json", jsonPatch)
}

func Kubectl(args ...string) *kubectl {
	return &kubectl{
		Name: "kubectl",
		Args: args,
	}
}
