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
	SmMarketplaceHelmChartName = "sm-marketplace"
	SmMarketplaceHelmRepoName  = "sm-marketplace"
	SmMarketplaceNamespace     = "sm-marketplace"
)

var (
	_ Resource = new(ServiceMeshHub)

	NoGithubTokenError = errors.Errorf("GITHUB_TOKEN not found")
)

type ServiceMeshHub struct {
	Version string
}

func (s *ServiceMeshHub) Teardown(ctx context.Context, command cmd.Factory) error {
	version, err := s.getVersion(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error determining version to install", zap.Error(err))
		return err
	}

	installed, err := smhInstalled(ctx, version)
	if err != nil {
		return err
	}
	if installed {
		return uninstallServiceMeshHub(ctx, command, version)
	}
	return nil
}

func (s *ServiceMeshHub) Ensure(ctx context.Context, command cmd.Factory) error {
	version, err := s.getVersion(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error determining version to install", zap.Error(err))
		return err
	}

	installed, err := smhInstalled(ctx, version)
	if err != nil {
		return err
	}
	if !installed {
		return installServiceMeshHub(ctx, command, version)
	}
	return nil
}

func smhInstalled(ctx context.Context, version string) (bool, error) {
	active, err := internal.NamespaceIsActive(ctx, SmMarketplaceNamespace)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error checking if namespace is active", zap.Error(err))
		return false, err
	}
	if !active {
		contextutils.LoggerFrom(ctx).Infow("sm-marketplace namespace does not exist.")
		return false, nil
	}

	ok, err := internal.PodsReadyAndVersionsMatch(ctx, SmMarketplaceNamespace, "app=sm-marketplace", version)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error checking pods and containers", zap.Error(err))
		return false, err
	}
	if !ok {
		contextutils.LoggerFrom(ctx).Infof("service mesh hub pods not running with expected version %s", version)
		return false, nil
	}
	contextutils.LoggerFrom(ctx).Infof("service mesh hub installed at desired version %s", version)
	return true, nil
}

func installServiceMeshHub(ctx context.Context, command cmd.Factory, version string) error {
	contextutils.LoggerFrom(ctx).Infof("preparing to install version %s", version)

	if err := addHelmRepo(ctx, command); err != nil {
		return err
	}

	untarDir, err := fetchAndUntarChart(ctx, command, version)
	if err != nil {
		return err
	}

	chartDir := filepath.Join(untarDir, SmMarketplaceHelmChartName)
	if err := renderAndApplyManifest(ctx, command, chartDir); err != nil {
		return err
	}

	return internal.WaitUntilPodsRunning(ctx, SmMarketplaceNamespace)
}

func renderAndApplyManifest(ctx context.Context, command cmd.Factory, chartDir string) error {
	contextutils.LoggerFrom(ctx).Infow("rendering and applying manifest for service mesh hub")
	out, err := command.Helm().Template().Namespace(SmMarketplaceNamespace).Set("namespace.create=true").Target(chartDir).Cmd().Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error rendering manifest", zap.Error(err), zap.String("out", out))
		return err
	}

	out, err = command.Kubectl().ApplyStdIn(out).Cmd().Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error applying manifest", zap.Error(err), zap.String("out", out))
	}
	return err
}

func uninstallServiceMeshHub(ctx context.Context, command cmd.Factory, version string) error {
	contextutils.LoggerFrom(ctx).Infof("preparing to uninstall version %s", version)

	if err := addHelmRepo(ctx, command); err != nil {
		return err
	}

	untarDir, err := fetchAndUntarChart(ctx, command, version)
	if err != nil {
		return err
	}

	chartDir := filepath.Join(untarDir, SmMarketplaceHelmChartName)
	if err := renderAndDeleteManifest(ctx, command, chartDir); err != nil {
		return err
	}

	return nil
}

func renderAndDeleteManifest(ctx context.Context, command cmd.Factory, chartDir string) error {
	contextutils.LoggerFrom(ctx).Infow("rendering and deleting manifest for service mesh hub")
	out, err := command.Helm().Template().Namespace(SmMarketplaceNamespace).Set("namespace.create=true").Target(chartDir).Cmd().Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error rendering manifest", zap.Error(err), zap.String("out", out))
		return err
	}

	out, err = command.Kubectl().DeleteStdIn(out).Cmd().Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error deleting manifest", zap.Error(err), zap.String("out", out))
	}
	return err
}

func addHelmRepo(ctx context.Context, command cmd.Factory) error {
	return command.Helm().AddRepo(SmMarketplaceHelmRepoName, SmMarketplaceHelmRepoUrl).Cmd().Run(ctx)
}

func fetchAndUntarChart(ctx context.Context, command cmd.Factory, version string) (string, error) {
	untarDir, err := getLocalDirectory(version)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error determining local directory for untarring chart", zap.Error(err))
		return "", err
	}
	if err := os.MkdirAll(untarDir, os.ModePerm); err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error making directory", zap.Error(err))
		return "", err
	}
	out, err := command.
		Helm().
		Fetch(SmMarketplaceHelmRepoName, SmMarketplaceHelmChartName).
		Version(version).
		UntarToDir(untarDir).
		Cmd().
		Output(ctx)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("error trying to untar helm chart", zap.Error(err), zap.String("out", out))
		return "", err
	}
	contextutils.LoggerFrom(ctx).Infow("successfully downloaded and extracted chart", zap.String("untarDir", untarDir))
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
