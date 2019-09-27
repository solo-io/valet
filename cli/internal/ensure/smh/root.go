package smh

import (
	"context"
	"fmt"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/valet/cli/api"
	"github.com/solo-io/valet/cli/internal"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

const (
	SmMarketplaceOwner = "solo-io"
	SmMarketplaceRepo = "sm-marketplace"
	SmMarketplaceHelmRepoUrl = "https://storage.googleapis.com/sm-marketplace-helm/"
	SmMarketplaceHelmRepoName = "sm-marketplace"
	SmMarketplaceHelmChartName = "sm-marketplace"
	SmMarketplaceNamespace = "sm-marketplace"
)

var (
	NoGithubTokenError = errors.Errorf("GITHUB_TOKEN not found")
)

func EnsureServiceMeshHub(ctx context.Context, serviceMeshHub *api.ServiceMeshHub) error {
	version, err := getVersion(ctx, serviceMeshHub)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error determining version to install", zap.Error(err))
		return err
	}
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
	out, err := internal.ExecuteCmd("helm", "template", "--namespace", SmMarketplaceNamespace, "--set", "namespace.create=true", chartDir)
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error rendering manifest", zap.Error(err), zap.String("out", out))
		return err
	}

	out, err = internal.ExecuteCmdStdIn(out, "kubectl", "apply", "-f", "-")
	if err != nil {
		contextutils.LoggerFrom(ctx).Errorw("Error applying manifest", zap.Error(err), zap.String("out", out))
	}
	return err
}

func addHelmRepo(ctx context.Context) error {
	contextutils.LoggerFrom(ctx).Infow("Adding helm repo")
	out, err := internal.ExecuteCmd("helm", "repo", "add", SmMarketplaceHelmRepoName, SmMarketplaceHelmRepoUrl)
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
	out, err := internal.ExecuteCmd("helm", "fetch",
		fmt.Sprintf("%s/%s", SmMarketplaceHelmRepoName, SmMarketplaceHelmChartName),
		"--version", version, "--untar", "--untardir", untarDir)
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

func getVersion(ctx context.Context, serviceMeshHub *api.ServiceMeshHub) (string, error) {
	if serviceMeshHub.Version != "" {
		return serviceMeshHub.Version, nil
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
