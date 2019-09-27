package gloo

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/api"
	"github.com/solo-io/valet/cli/internal"
	"go.uber.org/zap"
)

const (
	EnterpriseGlooRepo = "solo-projects"
	GlooRepo           = "gloo"
	DefaultNamespace   = "gloo-system"
	GlooSelector       = "gloo"
)

var (
	CouldNotDetermineVersionError = func(err error) error {
		return errors.Wrapf(err, "Error determining latest release.")
	}
	MustProvideVersionError = errors.Errorf("Must provide a version for install")

	_ GlooManager = new(glooManager)
)

type GlooManager interface {
	Install(ctx context.Context) error
	Glooctl() Glooctl
	Uninstall(ctx context.Context) error
}

func NewGlooManager(valet *api.Valet, gloo *api.Gloo) GlooManager {
	return &glooManager{
		valet: valet,
		gloo: gloo,
	}
}

type glooManager struct {
	valet *api.Valet
	gloo  *api.Gloo

	glooctl Glooctl
}

func (m *glooManager) Glooctl() Glooctl {
	return m.glooctl
}

func (m *glooManager) Install(ctx context.Context) error {
	if err := m.installGlooctl(ctx); err != nil {
		return err
	}
	version, err := m.getVersion(ctx)
	if err != nil {
		return err
	}
	glooInstalled, err := m.glooInstalled(ctx, version)
	if err != nil {
		return err
	}
	if !glooInstalled {
		return m.installGloo(ctx)
	}
	return nil
}

func (m *glooManager) getVersion(ctx context.Context) (string, error) {
	version := m.gloo.Version
	if version == "" {
		repo := GlooRepo
		if m.gloo.Enterprise {
			repo = EnterpriseGlooRepo
		}
		tag, err := getLatestTag(ctx, repo)
		if err != nil {
			return "", err
		}
		version = tag[1:]
	}
	return version, nil
}

func (m *glooManager) installGlooctl(ctx context.Context) error {
	m.glooctl = NewGlooctl(m.valet, m.gloo)
	return m.glooctl.Install(ctx)
}

func (m *glooManager) Uninstall(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Uninstalling existing gloo")
	out, err := m.glooctl.Execute("uninstall", "--all")
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to uninstall gloo",
			zap.Error(err),
			zap.String("out", out))
	}
	return err
}

func (m *glooManager) glooInstalled(ctx context.Context, version string) (bool, error) {
	active, err := internal.NamespaceIsActive(ctx, DefaultNamespace)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error checking if namespace is active", zap.Error(err))
		return false, err
	}
	if !active {
		contextutils.LoggerFrom(ctx).Infow("Gloo namespace does not exist.")
		return false, nil
	}
	if m.valet.LocalArtifactsDir != "" && m.gloo.Version == "" {
		// For local artifacts where we don't know the version, start with blank slate
		return false, m.Uninstall(ctx)
	}
	ok, err := podsReadyAndVersionsMatch(ctx, DefaultNamespace, GlooSelector, version)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error checking pods and containers", zap.Error(err))
		return false, err
	}
	if !ok {
		contextutils.LoggerFrom(ctx).Infow("Gloo pods not running with expected version, uninstalling")
		return false, m.Uninstall(ctx)
	}
	contextutils.LoggerFrom(ctx).Infow("Gloo installed at desired version")
	return true, nil
}

func (m *glooManager) installGloo(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Running glooctl install")
	args := []string{"install", "gateway"}
	if m.gloo.Enterprise {
		args = append(args, "enterprise", "--license-key", m.gloo.LicenseKey)
	}
	if m.valet.LocalArtifactsDir != "" && m.gloo.Version != "" {
		var helmChart string
		if m.gloo.Enterprise {
			helmChart = fmt.Sprintf("_artifacts/gloo-ee-%s.tgz", m.gloo.Version)
		} else {
			helmChart = fmt.Sprintf("_artifacts/gloo-%s.tgz", m.gloo.Version)
		}
		args = append(args, "-f", helmChart)
		contextutils.LoggerFrom(ctx).Infow("Using helm chart from local artifacts", zap.String("helmChart", helmChart))
	} else if m.valet.LocalArtifactsDir != ""{
		return MustProvideVersionError
	}
	if m.valet.ValetArtifacts && m.gloo.Version != "" {
		var helmChart string
		if m.gloo.Enterprise {
			helmChart = fmt.Sprintf("https://storage.googleapis.com/valet/artifacts/gloo/%s/gloo-%s.tgz", m.gloo.Version, m.gloo.Version)
		} else {
			helmChart = fmt.Sprintf("https://storage.googleapis.com/valet/artifacts/solo-projects/%s/gloo-ee-%s.tgz", m.gloo.Version, m.gloo.Version)
		}
		args = append(args, "-f", helmChart)
		contextutils.LoggerFrom(ctx).Infow("Using helm chart from valet artifacts", zap.String("helmChart", helmChart))
	} else if m.valet.ValetArtifacts {
		return MustProvideVersionError
	}

	if m.gloo.Version != "" && !m.valet.ValetArtifacts && m.valet.LocalArtifactsDir == "" {
		var helmChart string
		if m.gloo.Enterprise {
			helmChart = fmt.Sprintf("https://storage.googleapis.com/gloo-ee-helm/charts/gloo-ee-%s.tgz", m.gloo.Version)
		} else {
			helmChart = fmt.Sprintf("https://storage.googleapis.com/solo-public-helm/charts/gloo-%s.tgz", m.gloo.Version)
		}
		args = append(args, "-f", helmChart)
		contextutils.LoggerFrom(ctx).Infow("Using helm chart from release artifacts", zap.String("helmChart", helmChart))
	}
	out, err := m.glooctl.Execute(args...)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Failed to install gloo",
			zap.Error(err),
			zap.String("out", out))
		return err
	}
	return internal.WaitUntilPodsRunning(ctx, DefaultNamespace)
}


