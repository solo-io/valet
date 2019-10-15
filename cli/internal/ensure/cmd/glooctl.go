package cmd

import (
	"context"
	"strings"
)

type Glooctl Command

func (g *Glooctl) With(args ...string) *Glooctl {
	g.Args = append(g.Args, args...)
	return g
}

func (g *Glooctl) Command() *Command {
	return &Command{
		Name:            g.Name,
		Args:            g.Args,
		StdIn:           g.StdIn,
		Redactions:      g.Redactions,
		SwallowErrorLog: g.SwallowErrorLog,
	}
}

func (g *Glooctl) SwallowError() *Glooctl {
	g.SwallowErrorLog = true
	return g
}

func (g *Glooctl) Run(ctx context.Context) error {
	return g.Command().Run(ctx)
}

func (g *Glooctl) Output(ctx context.Context) (string, error) {
	return g.Command().Output(ctx)
}

func (g *Glooctl) UninstallAll() *Glooctl {
	return g.With("uninstall", "--all")
}

func (g *Glooctl) LicenseKey(licenseKey string) *Glooctl {
	return g.With("--license-key", licenseKey).Redact(licenseKey, Redacted)
}

func (g *Glooctl) Redact(unredacted, redacted string) *Glooctl {
	if g.Redactions == nil {
		g.Redactions = make(map[string]string)
	}
	g.Redactions[unredacted] = redacted
	return g
}

func (g *Glooctl) ProxyUrl() *Glooctl {
	return g.With("proxy", "url")
}

func (g *Glooctl) ProxyAddress() *Glooctl {
	return g.With("proxy", "url")
}

func (g *Glooctl) GetProxyIp(ctx context.Context) (string, error) {
	address, err := g.ProxyAddress().Output(ctx)
	if err != nil {
		return "", err
	}
	address = strings.TrimPrefix(address, "http://")
	address = strings.TrimPrefix(address, "https://")
	portIndex := strings.Index(address, ":")
	if portIndex >= 0 {
		address = address[:portIndex]
	}
	return address, nil
}

func (g *Glooctl) GetUpstream(name string) *Glooctl {
	return g.With("get", "upstream", name)
}

func (g *Glooctl) CreateUpstream(name string) *Glooctl {
	return g.With("create", "upstream", name)
}

func (g *Glooctl) AwsSecretName(secretName string) *Glooctl {
	return g.With("--aws-secret-name", secretName)
}
