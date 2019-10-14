package resource

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal"
	"go.uber.org/zap"
)

const (
	GlooRepo           = "gloo"
	GlooEnterpriseRepo = "solo-projects"
	DefaultNamespace   = "gloo-system"
	GlooSelector       = "gloo"
	AwsSecretName      = "aws-creds"
	AwsUpstreamName    = "aws"
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

	glooctl *Glooctl
}

func (g *Gloo) Ensure(ctx context.Context) error {
	glooctl := g.getGlooctl(ctx)
	if err := glooctl.Ensure(ctx); err != nil {
		return err
	}
	g.glooctl = glooctl
	version, err := g.getVersion(ctx)
	if err != nil {
		return err
	}
	glooInstalled, err := g.glooInstalled(ctx, version)
	if err != nil {
		return err
	}
	if glooInstalled {
		return nil
	}

	err = g.installGloo(ctx)
	if err != nil {
		return err
	}

	if g.UiVirtualService != nil {
		if g.UiVirtualService.DNS != nil {
			proxyIp, err := g.glooctl.GetProxyIp(ctx)
			if err != nil {
				return err
			}
			g.UiVirtualService.DNS.IP = proxyIp
		}
	}

	g.AWS.Glooctl = g.glooctl
	if err = g.AWS.Ensure(ctx); err != nil {
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

func (g *Gloo) glooInstalled(ctx context.Context, version string) (bool, error) {
	active, err := internal.NamespaceIsActive(ctx, DefaultNamespace)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error checking if namespace is active", zap.Error(err))
		return false, err
	}
	if !active {
		contextutils.LoggerFrom(ctx).Infow("Gloo namespace does not exist.")
		return false, nil
	}
	if g.LocalArtifactsDir != "" && g.Version == "" {
		// For local artifacts where we don't know the version, start with blank slate
		return false, g.uninstall(ctx)
	}
	ok, err := internal.PodsReadyAndVersionsMatch(ctx, DefaultNamespace, GlooSelector, version)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error checking pods and containers", zap.Error(err))
		return false, err
	}
	if !ok {
		contextutils.LoggerFrom(ctx).Infow("Gloo pods not running with expected version, uninstalling")
		return false, g.uninstall(ctx)
	}
	contextutils.LoggerFrom(ctx).Infow("Gloo installed at desired version")
	return true, nil
}

func (g *Gloo) installGloo(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Running glooctl install")
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
		contextutils.LoggerFrom(ctx).Infow("Using helm chart from local artifacts", zap.String("helmChart", helmChart))
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
		contextutils.LoggerFrom(ctx).Infow("Using helm chart from valet artifacts", zap.String("helmChart", helmChart))
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
		contextutils.LoggerFrom(ctx).Infow("Using helm chart from release artifacts", zap.String("helmChart", helmChart))
	}
	glooctlCmd, err := g.glooctl.Command()
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to construct glooctlCmd", zap.Error(err))
		return err
	}
	out, err := glooctlCmd.Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to install gloo",
			zap.Error(err),
			zap.String("out", out))
		return err
	}
	return internal.WaitUntilPodsRunning(ctx, DefaultNamespace)
}

func (g *Gloo) uninstall(ctx context.Context) error {
	if g.glooctl == nil {
		return GlooctlNotEnsuredError
	}
	glooctlCmd, err := g.glooctl.Command()
	if err != nil {
		return err
	}
	return glooctlCmd.UninstallAll().Run(ctx)
}

func (g *Gloo) Teardown(ctx context.Context) error {
	if g.glooctl == nil {
		g.glooctl = g.getGlooctl(ctx)
		if err := g.glooctl.Ensure(ctx); err != nil {
			return err
		}
	}
	return g.uninstall(ctx)
}

func (g *Gloo) GetProxyIp(ctx context.Context) (string, error) {
	if g.glooctl == nil {
		g.glooctl = g.getGlooctl(ctx)
		if err := g.glooctl.Ensure(ctx); err != nil {
			return "", err
		}
	}
	return g.glooctl.GetProxyIp(ctx)
}

func (g *Gloo) GetProxyAddress(ctx context.Context) (string, error) {
	if g.glooctl == nil {
		g.glooctl = g.getGlooctl(ctx)
		if err := g.glooctl.Ensure(ctx); err != nil {
			return "", err
		}
	}
	return g.glooctl.GetProxyAddress(ctx)
}

type UiVirtualService struct {
	// If nil, the default domain "*" is used. Otherwise, a DNS entry is created in Route53
	// with the provided DNS configuration.
	DNS *DNS `yaml:"dns"`
}

func (u *UiVirtualService) Ensure(ctx context.Context) error {
	if u.DNS != nil {
		if err := u.DNS.Ensure(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (u *UiVirtualService) Teardown(ctx context.Context) error {
	if u.DNS != nil {
		if err := u.DNS.Teardown(ctx); err != nil {
			return err
		}
	}
	return nil
}

type AWS struct {
	Secret   bool `yaml:"secret"`
	Upstream bool `yaml:"upstream"`

	Glooctl *Glooctl
}

func (a *AWS) Ensure(ctx context.Context) error {
	if a.Secret {
		secret := AwsSecret(AwsSecretName, DefaultNamespace)
		if err := secret.Ensure(ctx); err != nil {
			return err
		}
	}
	if a.Upstream {
		if a.Glooctl == nil {
			return GlooctlNotProvidedError
		}
		glooctlCmd, err := a.Glooctl.Command()
		if err != nil {
			return err
		}
		err = glooctlCmd.GetUpstream(AwsUpstreamName).Run(ctx)
		if err != nil {
			err = glooctlCmd.CreateUpstream(AwsUpstreamName).AwsSecretName(AwsSecretName).Run(ctx)
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
