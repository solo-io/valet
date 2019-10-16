package resource

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"go.uber.org/zap"
	"os"
)

const (
	GlooRepo           = "gloo"
	GlooEnterpriseRepo = "solo-projects"
	DefaultNamespace   = "gloo-system"
	GlooSelector       = "gloo"
	AwsSecretName      = "aws-creds"
	AwsUpstreamName    = "aws"
	LicenseKeyEnvVar   = "LICENSE_KEY"
)

var (
	_ Resource = new(Gloo)

	MustProvideVersionError = errors.Errorf("must provide a version for install")
)

type Gloo struct {
	// Gloo (server) version. This should not begin with a "v".
	Version           string            `yaml:"version"`
	Enterprise        bool              `yaml:"enterprise"`
	ValetArtifacts    bool              `yaml:"valetArtifacts"`
	LocalArtifactsDir string            `yaml:"localArtifactsDir"`
	LicenseKey        string            `yaml:"licenseKey"`
	UiVirtualService  *UiVirtualService `yaml:"uiVirtualService"`
	AWS               AWS               `yaml:"aws"`
}

func (g *Gloo) Ensure(ctx context.Context, command cmd.Factory) error {
	glooctl := g.getGlooctl(ctx)
	if err := glooctl.Ensure(ctx, command); err != nil {
		return err
	}
	version, err := g.getVersion(ctx)
	if err != nil {
		return err
	}
	glooInstalled, err := g.glooInstalled(ctx, command, version)
	if err != nil {
		return err
	}
	if !glooInstalled {
		if g.LicenseKey == "" {
			g.LicenseKey = os.Getenv(LicenseKeyEnvVar)
		}

		err = g.installGloo(ctx, command)
		if err != nil {
			return err
		}
	}

	if g.UiVirtualService != nil {
		if g.UiVirtualService.DNS != nil {
			proxyIp, err := command.Glooctl().GetProxyIp(ctx)
			if err != nil {
				return err
			}
			g.UiVirtualService.DNS.IP = proxyIp
		}
	}

	if err = g.AWS.Ensure(ctx, command); err != nil {
		return err
	}

	return nil
}

func (g *Gloo) getGlooctl(ctx context.Context) *Glooctl {
	return &Glooctl{
		Enterprise:        g.Enterprise,
		Version:           g.Version,
		LocalArtifactsDir: g.LocalArtifactsDir,
		ValetArtifacts:    g.ValetArtifacts,
	}
}

func (g *Gloo) getVersion(ctx context.Context) (string, error) {
	version := g.Version
	if version == "" {
		repo := GlooRepo
		if g.Enterprise {
			repo = GlooEnterpriseRepo
		}
		tag, err := getLatestTag(ctx, repo)
		if err != nil {
			return "", err
		}
		version = tag[1:]
	}
	return version, nil
}

func (g *Gloo) glooInstalled(ctx context.Context, command cmd.Factory, version string) (bool, error) {
	active, err := internal.NamespaceIsActive(ctx, DefaultNamespace)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error checking if namespace is active", zap.Error(err))
		return false, err
	}
	if !active {
		contextutils.LoggerFrom(ctx).Infow("gloo namespace does not exist.")
		return false, nil
	}
	if g.LocalArtifactsDir != "" && g.Version == "" {
		// For local artifacts where we don't know the version, start with blank slate
		return false, g.uninstall(ctx, command)
	}
	ok, err := internal.PodsReadyAndVersionsMatch(ctx, DefaultNamespace, GlooSelector, version)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error checking pods and containers", zap.Error(err))
		return false, err
	}
	if !ok {
		contextutils.LoggerFrom(ctx).Infow("uninstalling existing gloo")
		return false, g.uninstall(ctx, command)
	}
	contextutils.LoggerFrom(ctx).Infow("gloo installed at desired version")
	return true, nil
}

func (g *Gloo) installGloo(ctx context.Context, command cmd.Factory) error {
	args := []string{"install", "gateway"}
	if g.Enterprise {
		args = append(args, "enterprise", "--license-key", g.LicenseKey)
	}
	if g.LocalArtifactsDir != "" && g.Version != "" {
		var helmChart string
		if g.Enterprise {
			helmChart = fmt.Sprintf("_artifacts/gloo-ee-%s.tgz", g.Version)
		} else {
			helmChart = fmt.Sprintf("_artifacts/gloo-%s.tgz", g.Version)
		}
		args = append(args, "-f", helmChart)
	} else if g.LocalArtifactsDir != "" {
		return MustProvideVersionError
	}
	if g.ValetArtifacts && g.Version != "" {
		var helmChart string
		if g.Enterprise {
			helmChart = fmt.Sprintf("https://storage.googleapis.com/valet/artifacts/gloo/%s/gloo-%s.tgz", g.Version, g.Version)
		} else {
			helmChart = fmt.Sprintf("https://storage.googleapis.com/valet/artifacts/solo-projects/%s/gloo-ee-%s.tgz", g.Version, g.Version)
		}
		args = append(args, "-f", helmChart)
	} else if g.ValetArtifacts {
		return MustProvideVersionError
	}

	if g.Version != "" && !g.ValetArtifacts && g.LocalArtifactsDir == "" {
		var helmChart string
		if g.Enterprise {
			helmChart = fmt.Sprintf("https://storage.googleapis.com/gloo-ee-helm/charts/gloo-ee-%s.tgz", g.Version)
		} else {
			helmChart = fmt.Sprintf("https://storage.googleapis.com/solo-public-helm/charts/gloo-%s.tgz", g.Version)
		}
		args = append(args, "-f", helmChart)
	}
	out, err := command.Glooctl().With(args...).Redact(g.LicenseKey, cmd.Redacted).Cmd().Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("failed to install gloo",
			zap.Error(err),
			zap.String("out", out))
		return err
	}
	return internal.WaitUntilPodsRunning(ctx, DefaultNamespace)
}

func (g *Gloo) uninstall(ctx context.Context, command cmd.Factory) error {
	if command.Glooctl() == nil {
		return GlooctlNotEnsuredError
	}
	return command.Glooctl().UninstallAll().Cmd().Run(ctx)
}

func (g *Gloo) Teardown(ctx context.Context, command cmd.Factory) error {
	glooctl := g.getGlooctl(ctx)
	if err := glooctl.Ensure(ctx, command); err != nil {
		return err
	}
	return g.uninstall(ctx, command)
}

func (g *Gloo) GetProxyIp(ctx context.Context, command cmd.Factory) (string, error) {
	if command.Glooctl() == nil {
		glooctl := g.getGlooctl(ctx)
		if err := glooctl.Ensure(ctx, command); err != nil {
			return "", err
		}
	}
	return command.Glooctl().GetProxyIp(ctx)
}

func (g *Gloo) GetProxyAddress(ctx context.Context, command cmd.Factory) (string, error) {
	if command.Glooctl() == nil {
		glooctl := g.getGlooctl(ctx)
		if err := glooctl.Ensure(ctx, command); err != nil {
			return "", err
		}
	}
	return command.Glooctl().ProxyAddress().Cmd().Output(ctx)
}

const GlooUiVirtualService = `
apiVersion: gateway.solo.io/v1
kind: VirtualService
metadata:
  name: glooui
  namespace: gloo-system
spec:
  virtualHost:
    domains:
      - "*"
    routes:
      - matcher:
          prefix: /
        routeAction:
          single:
            upstream:
              name: gloo-system-apiserver-ui-8080
              namespace: gloo-system`

type UiVirtualService struct {
	// If nil, the default domain "*" is used. Otherwise, a DNS entry is created in Route53
	// with the provided DNS configuration.
	DNS *DNS `yaml:"dns"`
}

func (u *UiVirtualService) Ensure(ctx context.Context, command cmd.Factory) error {
	if err := command.Kubectl().ApplyStdIn(GlooUiVirtualService).Cmd().Run(ctx); err != nil {
		return err
	}
	if u.DNS != nil {
		if err := u.DNS.Ensure(ctx, command); err != nil {
			return err
		}
	}

	if err := patchGloouiWithDomain(ctx, command, u.DNS.Domain); err != nil {
		return err
	}

	if u.DNS.Cert != nil {
		if err := patchGloouiWithSsl(ctx, command, u.DNS.Domain); err != nil {
			return err
		}
	}

	return nil
}

func (u *UiVirtualService) Teardown(ctx context.Context, command cmd.Factory) error {
	if u.DNS != nil {
		if err := u.DNS.Teardown(ctx, command); err != nil {
			return err
		}
	}
	return nil
}

type AWS struct {
	Secret   bool `yaml:"secret"`
	Upstream bool `yaml:"upstream"`
}

func (a *AWS) Ensure(ctx context.Context, command cmd.Factory) error {
	if a.Secret {
		secret := AwsSecret(DefaultNamespace, AwsSecretName)
		if err := secret.Ensure(ctx, command); err != nil {
			return err
		}
	}
	if a.Upstream {
		if command.Glooctl() == nil {
			return GlooctlNotProvidedError
		}
		err := command.Glooctl().GetUpstream(AwsUpstreamName).SwallowError().Cmd().Run(ctx)
		if err != nil {
			err = command.Glooctl().CreateUpstream(AwsUpstreamName).AwsSecretName(AwsSecretName).With("--name", AwsUpstreamName).Cmd().Run(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *AWS) Teardown(ctx context.Context) error {
	panic("implement me")
}

func patchGloouiWithDomain(ctx context.Context, command cmd.Factory, domain string) error {
	contextutils.LoggerFrom(ctx).Infow("Patching glooui domain")
	patchStr := fmt.Sprintf("-p=[{\"op\":\"add\",\"path\":\"/spec/virtualHost/domains\",\"value\":[\"%s\"]}]", domain)
	return command.Kubectl().With("patch", "vs", "glooui").Namespace("gloo-system").JsonPatch(patchStr).Cmd().Run(ctx)
}

func patchGloouiWithSsl(ctx context.Context, command cmd.Factory, domain string) error {
	contextutils.LoggerFrom(ctx).Infow("Patching glooui ssl config")
	patchStr := fmt.Sprintf(`spec:
  sslConfig:
    secretRef:
      name: %s
      namespace: gloo-system
    sniDomains:
    - %s`, domain, domain)
	return command.Kubectl().With("patch", "vs", "glooui").Namespace("gloo-system").With("--patch", patchStr, "--type=merge").Cmd().Run(ctx)
}
