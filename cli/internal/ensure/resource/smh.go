package resource

import (
	"context"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/valet/cli/internal"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

const (
	SmMarketplaceOwner         = "solo-io"
	SmMarketplaceRepo          = "sm-marketplace"
	SmMarketplaceHelmRepoUrl   = "https://storage.googleapis.com/sm-marketplace-helm/"
	SmMarketplaceHelmRepoName  = "sm-marketplace"
	SmMarketplaceHelmChartName = "sm-marketplace"
	SmMarketplaceNamespace     = "sm-marketplace"
)

var (
	_ Resource = new(ServiceMeshHub)

	NoGithubTokenError = errors.Errorf("GITHUB_TOKEN not found")
)

type ServiceMeshHub struct {
	Version string
}

func (s *ServiceMeshHub) Teardown(ctx context.Context) error {
	return nil
}

func (s *ServiceMeshHub) Ensure(ctx context.Context) error {
	version, err := s.getVersion(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error determining version to install", zap.Error(err))
		return err
	}

	installed, err := smhInstalled(ctx, version)
	if err != nil {
		return err
	}
	if !installed {
		return installServiceMeshHub(ctx, version)
	}
	return nil
}

func smhInstalled(ctx context.Context, version string) (bool, error) {
	active, err := internal.NamespaceIsActive(ctx, SmMarketplaceNamespace)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error checking if namespace is active", zap.Error(err))
		return false, err
	}
	if !active {
		contextutils.LoggerFrom(ctx).Infow("sm-marketplace namespace does not exist.")
		return false, nil
	}

	ok, err := internal.PodsReadyAndVersionsMatch(ctx, SmMarketplaceNamespace, "app=sm-marketplace", version)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error checking pods and containers", zap.Error(err))
		return false, err
	}
	if !ok {
		contextutils.LoggerFrom(ctx).Infow("service mesh hub pods not running with expected version")
		return false, nil
	}
	contextutils.LoggerFrom(ctx).Infow("service mesh hub installed at desired version")
	return true, nil
}

func installServiceMeshHub(ctx context.Context, version string) error {
	contextutils.LoggerFrom(ctx).Infof("Preparing to install version %s", version)

	if err := addHelmRepo(ctx); err != nil {
		return err
	}

	untarDir, err := fetchAndUntarChart(ctx, version)
	if err != nil {
		return err
	}

	chartDir := filepath.Join(untarDir, SmMarketplaceHelmChartName)
	if err := renderAndApplyManifest(ctx, chartDir); err != nil {
		return err
	}

	return internal.WaitUntilPodsRunning(ctx, SmMarketplaceNamespace)
}

func renderAndApplyManifest(ctx context.Context, chartDir string) error {
	contextutils.LoggerFrom(ctx).Infow("Rendering and applying manifest for service mesh hub")
	// helm template --set namespace.create=true ~/.helm/untar/sm-marketplace/0.2.3/sm-marketplace/
	out, err := cmd.Helm().Template().Namespace(SmMarketplaceNamespace).Set("namespace.create=true").Target(chartDir).Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error rendering manifest", zap.Error(err), zap.String("out", out))
		return err
	}

	out, err = cmd.Kubectl().ApplyStdIn(out).Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error applying manifest", zap.Error(err), zap.String("out", out))
	}
	return err
}

func addHelmRepo(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Adding helm repo")
	out, err := cmd.Helm().AddRepo(SmMarketplaceHelmRepoName, SmMarketplaceHelmRepoUrl).Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error trying to add repo", zap.Error(err), zap.String("out", out))
	}
	return err
}

func fetchAndUntarChart(ctx context.Context, version string) (string, error) {
	// helm fetch sm-marketplace/sm-marketplace --version 0.2.3 --untar -untardir ~/.helm/untar/sm-marketplace/0.2.3
	untarDir, err := getLocalDirectory(version)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error determining local directory for untarring chart", zap.Error(err))
		return "", err
	}
	if err := os.MkdirAll(untarDir, os.ModePerm); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error making directory", zap.Error(err))
		return "", err
	}
	out, err := cmd.
		Helm().
		Fetch(SmMarketplaceHelmRepoName, SmMarketplaceHelmRepoUrl).
		Version(version).
		UntarToDir(untarDir).
		Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error trying to untar helm chart", zap.Error(err), zap.String("out", out))
		return "", err
	}
	contextutils.LoggerFrom(ctx).Infow("Successfully downloaded and extracted chart", zap.String("untarDir", untarDir))
	return untarDir, nil
}

func getLocalDirectory(version string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".helm", "untar", SmMarketplaceHelmRepoName, SmMarketplaceHelmChartName, version), nil
}

func (s *ServiceMeshHub) getVersion(ctx context.Context) (string, error) {
	if s.Version != "" {
		return s.Version, nil
	}
	contextutils.LoggerFrom(ctx).Infow("Determining version to install for service mesh hub")
	if os.Getenv(githubutils.GITHUB_TOKEN) == "" {
		return "", NoGithubTokenError
	}
	client, err := githubutils.GetClient(ctx)
	if err != nil {
		return "", err
	}
	latestTag, err := githubutils.FindLatestReleaseTag(ctx, client, SmMarketplaceOwner, SmMarketplaceRepo)
	if err != nil {
		return "", err
	}
	return latestTag[1:], nil
}
